// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"time"

	"api/internal/svc"
	"api/internal/types"
	"api/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListDailyTasksLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListDailyTasksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListDailyTasksLogic {
	return &ListDailyTasksLogic{
		Logger: logWithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func logWithContext(ctx context.Context) logx.Logger {
	return logx.WithContext(ctx)
}

func (l *ListDailyTasksLogic) ListDailyTasks(req *types.DailyTaskQueryReq) (resp *types.DailyTaskListResp, err error) {
	if l.svcCtx.DailyTasksModel == nil {
		return okDailyTaskList(), nil
	}

	query, args := buildDailyTaskListQuery(currentUserID(l.ctx), req)

	var items []model.DailyTasks
	if err = l.svcCtx.DB.QueryRowsCtx(l.ctx, &items, query, args...); err != nil {
		return nil, fmt.Errorf("query daily tasks: %w", err)
	}

	// 获取当前用户的意识cycle配置，用于处理遗漏日期
	var startDate, endDate time.Time
	if req.StartDate != "" {
		startDate, _ = time.Parse("2006-01-02", req.StartDate)
	}
	if req.EndDate != "" {
		endDate, _ = time.Parse("2006-01-02", req.EndDate)
	}

	// 构建已有任务的日期映射
	existingDates := make(map[string]bool)
	list := make([]types.DailyTaskInfo, 0, len(items))

	for i := range items {
		taskDateStr := items[i].TaskDate.Format("2006-01-02")
		existingDates[taskDateStr] = true
		list = append(list, dailyTaskToInfo(&items[i]))
	}

	// 处理有缺失日期的情况 - 为遗漏的日期填补任务
	if startDate.IsZero() || endDate.IsZero() || startDate.After(endDate) {
		// 如果没有明确的日期范围，跳过自动生成
		l.Logger.Info("skipping auto-fill for missing dates: no valid date range")
	} else {
		// 尝试为缺失的日期生成任务
		l.fillMissingDates(startDate, endDate, existingDates, &list)
	}

	// 按taskDate降序排序（最新的在前）
	sortDailyTasksByDate(list)

	return &types.DailyTaskListResp{
		Code:    0,
		Message: "ok",
		Data: types.DailyTaskListData{
			List:       list,
			Pagination: types.Pagination{Page: 1, PageSize: int64(len(list)), Total: int64(len(list))},
		},
	}, nil
}

// fillMissingDates 为指定日期范围内缺失的日期自动生成任务记录（支持补卡）
func (l *ListDailyTasksLogic) fillMissingDates(startDate, endDate time.Time, existingDates map[string]bool, list *[]types.DailyTaskInfo) {
	if l.svcCtx.DailyTasksModel == nil || l.svcCtx.AwarenessModel == nil {
		return
	}

	userID := currentUserID(l.ctx)
	now := time.Now()

	// 获取有效的意识点列表
	points, err := l.svcCtx.AwarenessModel.FindEligible(l.ctx)
	if err != nil {
		l.Logger.Infof("skipping auto-fill: failed to find eligible awareness points: %v", err)
		return
	}

	// 获取cycle设置
	startSetting, restDays, err := getAwarenessCycleSettings(l.ctx, l.svcCtx)
	if err != nil {
		l.Logger.Infof("skipping auto-fill: failed to get awareness cycle settings: %v", err)
		return
	}

	totalPoints := len(points)
	if totalPoints == 0 {
		// 没有可用意识点时，只生成占位符任务
		l.generatePlaceholderTasks(startDate, endDate, existingDates, userID, list)
		return
	}

	cycleLength := totalPoints + restDays

	// 遍历日期范围内的每一天
	for date := startDate; !date.After(endDate); date = date.AddDate(0, 0, 1) {
		dateStr := date.Format("2006-01-02")

		// 检查日期是否已有任务记录
		if existingDates[dateStr] {
			continue
		}

		// 检查此日期的cycle位置
		daysSinceStart := int(date.Sub(startSetting).Hours() / 24)
		if daysSinceStart < 0 {
			// 在cycle开始之前，生成占位符任务
			newItem := createMissedTaskInfo(date, "pre-start", "", "", "待开始", now)
			*list = append(*list, newItem)
			continue
		}

		dayInCycle := daysSinceStart % cycleLength
		if dayInCycle >= totalPoints {
			// 这是休息日，也生成占位符
			restDayNum := dayInCycle - totalPoints + 1
			restLabel := fmt.Sprintf("休息日 %d/%d", restDayNum, restDays)
			newItem := createMissedTaskInfo(date, "rest", "", "", restLabel, now)
			*list = append(*list, newItem)
			continue
		}

		// 这是一个正常训练日，应该生成任务
		awareness := &points[dayInCycle]

		// 为用户创建任务记录（在数据库中）
		userIDVal := currentUserID(l.ctx)
		dbItem := &model.DailyTasks{
			UserId:       userIDVal,
			CommunityId:  defaultCommunityID,
			TaskDate:     date,
			TopicId:      0,
			AwarenessId: sql.NullInt64{Int64: int64(awareness.AwarenessId), Valid: true},
			TopicOrderNo: awareness.SortOrderGlobal,
			TopicTitle:   awareness.PointTitle,
			TopicSummary: awarenessSummary(awareness),
			Status:       "draft", // 草稿状态，表示这是补卡任务
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		insertResult, insertErr := l.svcCtx.DailyTasksModel.Insert(l.ctx, dbItem)
		if insertErr != nil {
			l.Logger.Infof("failed to create missed task for %s: %v, adding in-memory entry", dateStr, insertErr)
			// 即使插入失败，也生成一个内存中的条目供用户查看
			newItem := createMissedTaskInfo(date, "draft", awareness.PointTitle, awarenessSummary(awareness), "可补卡", now)
			*list = append(*list, newItem)
			continue
		}

		// 获取新创建的任务ID
		newID, _ := insertResult.LastInsertId()

		// 构建任务信息返回给前端
		taskInfo := types.DailyTaskInfo{
			Id:                  uint64(newID),
			TaskDate:            dateStr,
			TopicId:             0,
			TopicOrderNo:        awareness.SortOrderGlobal,
			TopicTitle:          awareness.PointTitle,
			TopicSummary:       awarenessSummary(&points[dayInCycle]),
			TopicDescription:   "",
			Weakness:            "",
			ImprovementPlan:    "",
			VerificationPath:   "",
			ReflectionNote:    "",
			Status:             "draft",
			CanEditContent:     true, // 补卡任务在72小时内可以编辑
			CanAppendReflection: false,
			SubmittedAt:        "",
			CreatedAt:           now.Format("2006-01-02 15:04:05"),
			UpdatedAt:          now.Format("2006-01-02 15:04:05"),
		}
		*list = append(*list, taskInfo)
		existingDates[dateStr] = true
	}
}

// generatePlaceholderTasks 为日期范围内没有可用意识点时生成占位符任务
func (l *ListDailyTasksLogic) generatePlaceholderTasks(startDate, endDate time.Time, existingDates map[string]bool, userID uint64, list *[]types.DailyTaskInfo) {
	now := time.Now()
	for date := startDate; !date.After(endDate); date = date.AddDate(0, 0, 1) {
		dateStr := date.Format("2006-01-02")
		if !existingDates[dateStr] {
			newItem := createMissedTaskInfo(date, "pause", "", "", "休息日", now)
			*list = append(*list, newItem)
		}
	}
}

// createMissedTaskInfo 创建补卡任务的信息结构
func createMissedTaskInfo(taskDate time.Time, status, title, summary, desc string, now time.Time) types.DailyTaskInfo {
	dateStr := taskDate.Format("2006-01-02")
	info := types.DailyTaskInfo{
		Id:                  0,
		TaskDate:            dateStr,
		TopicId:             0,
		TopicOrderNo:        0,
		TopicTitle:          title,
		TopicSummary:        summary,
		TopicDescription:    "",
		Weakness:            "",
		ImprovementPlan:    "",
		VerificationPath:   "",
		ReflectionNote:    "",
		Status:             status,
		CanEditContent:     status == "draft",
		CanAppendReflection: false,
		SubmittedAt:        "",
		CreatedAt:           now.Format("2006-01-02 15:04:05"),
		UpdatedAt:          now.Format("2006-01-02 15:04:05"),
	}
	return info
}

// sortDailyTasksByDate 按taskDate降序排序（最新的在前）
func sortDailyTasksByDate(list []types.DailyTaskInfo) {
	sort.Slice(list, func(i, j int) bool {
		return list[i].TaskDate > list[j].TaskDate
	})
}

func buildDailyTaskListQuery(userID uint64, req *types.DailyTaskQueryReq) (string, []any) {
	query := "select id, user_id, community_id, task_date, schedule_day_id, topic_id, awareness_id, topic_order_no, topic_title, topic_summary, weakness, improvement_plan, verification_path, reflection_note, status, submitted_at, created_at, updated_at from daily_tasks where user_id = ?"
	args := []any{userID}

	if req.Status != "" {
		query += " and status = ?"
		args = append(args, req.Status)
	}

	if req.StartDate != "" {
		query += " and task_date >= ?"
		args = append(args, req.StartDate)
	}

	if req.EndDate != "" {
		query += " and task_date <= ?"
		args = append(args, req.EndDate)
	}

	if req.Keyword != "" {
		query += " and (topic_title like ? or weakness like ? or improvement_plan like ? or verification_path like ? or reflection_note like ?)"
		keyword := "%" + req.Keyword + "%"
		args = append(args, keyword, keyword, keyword, keyword, keyword)
	}

	query += " order by task_date desc limit 100"
	return query, args
}
