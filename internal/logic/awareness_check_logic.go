package logic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"time"

	"api/internal/svc"
	"api/internal/types"
	"api/model"

	"github.com/go-sql-driver/mysql"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	checkStatusDraft      = "draft"
	checkStatusInProgress = "in_progress"
	checkStatusAbandoned  = "abandoned"

	checkChapterStatusNotStarted = "not_started"
	checkChapterStatusInProgress = "in_progress"
)

type awarenessCheckSources struct {
	chapters        []model.Chapter
	points          []model.Awareness
	chapterByID     map[uint64]model.Chapter
	pointsByChapter map[uint64][]model.Awareness
	pointByID       map[uint64]model.Awareness
}

type GetAwarenessCheckCurrentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAwarenessCheckCurrentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAwarenessCheckCurrentLogic {
	return &GetAwarenessCheckCurrentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAwarenessCheckCurrentLogic) GetAwarenessCheckCurrent() (*types.AwarenessCheckCurrentResp, error) {
	if !hasAwarenessCheckStorage(l.svcCtx) {
		return &types.AwarenessCheckCurrentResp{Code: 0, Message: "ok", Data: types.AwarenessCheckCurrentData{}}, nil
	}

	check, err := loadOrCreateCurrentAwarenessCheck(l.ctx, l.svcCtx, currentUserID(l.ctx))
	if err != nil {
		return nil, err
	}
	data, err := buildAwarenessCheckCurrentData(l.ctx, l.svcCtx, check)
	if err != nil {
		return nil, err
	}
	return &types.AwarenessCheckCurrentResp{Code: 0, Message: "ok", Data: data}, nil
}

type CreateAwarenessCheckLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateAwarenessCheckLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateAwarenessCheckLogic {
	return &CreateAwarenessCheckLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateAwarenessCheckLogic) CreateAwarenessCheck() (*types.AwarenessCheckCurrentResp, error) {
	if !hasAwarenessCheckStorage(l.svcCtx) {
		return &types.AwarenessCheckCurrentResp{Code: 0, Message: "ok", Data: types.AwarenessCheckCurrentData{}}, nil
	}

	userID := currentUserID(l.ctx)
	sources, err := loadAwarenessCheckSources(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}
	if err = abandonCurrentAwarenessChecks(l.ctx, l.svcCtx, userID); err != nil {
		return nil, err
	}
	check, err := createNewAwarenessCheck(l.ctx, l.svcCtx, userID, sources)
	if err != nil {
		return nil, err
	}
	data, err := buildAwarenessCheckCurrentData(l.ctx, l.svcCtx, check)
	if err != nil {
		return nil, err
	}
	return &types.AwarenessCheckCurrentResp{Code: 0, Message: "ok", Data: data}, nil
}

type GetAwarenessCheckChapterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAwarenessCheckChapterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAwarenessCheckChapterLogic {
	return &GetAwarenessCheckChapterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAwarenessCheckChapterLogic) GetAwarenessCheckChapter(chapterID uint64) (*types.AwarenessCheckChapterResp, error) {
	if !hasAwarenessCheckStorage(l.svcCtx) {
		return &types.AwarenessCheckChapterResp{Code: 0, Message: "ok", Data: types.AwarenessCheckChapterData{}}, nil
	}

	check, err := loadOrCreateCurrentAwarenessCheck(l.ctx, l.svcCtx, currentUserID(l.ctx))
	if err != nil {
		return nil, err
	}
	data, err := buildAwarenessCheckChapterData(l.ctx, l.svcCtx, check, chapterID)
	if err != nil {
		return nil, err
	}
	return &types.AwarenessCheckChapterResp{Code: 0, Message: "ok", Data: data}, nil
}

type SaveAwarenessCheckChapterScoresLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSaveAwarenessCheckChapterScoresLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveAwarenessCheckChapterScoresLogic {
	return &SaveAwarenessCheckChapterScoresLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SaveAwarenessCheckChapterScoresLogic) SaveAwarenessCheckChapterScores(chapterID uint64, req *types.AwarenessCheckScoreSaveReq) (*types.AwarenessCheckChapterResp, error) {
	if !hasAwarenessCheckStorage(l.svcCtx) {
		return &types.AwarenessCheckChapterResp{Code: 0, Message: "ok", Data: types.AwarenessCheckChapterData{}}, nil
	}
	if req == nil || len(req.Scores) == 0 {
		return nil, fmt.Errorf("scores required")
	}

	userID := currentUserID(l.ctx)
	check, err := loadOrCreateCurrentAwarenessCheck(l.ctx, l.svcCtx, userID)
	if err != nil {
		return nil, err
	}
	chapter, err := l.svcCtx.AwarenessCheckChaptersModel.FindOneByCheckAndChapter(l.ctx, check.CheckId, chapterID)
	if err != nil {
		return nil, fmt.Errorf("query awareness check chapter: %w", err)
	}
	if chapter.UserId != userID {
		return nil, fmt.Errorf("awareness check chapter not found")
	}
	sources, err := loadAwarenessCheckSources(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}
	if chapter.Status == checkStatusCompleted {
		check, err = startNewChapterAttempt(l.ctx, l.svcCtx, check, sources)
		if err != nil {
			return nil, err
		}
		chapter, err = l.svcCtx.AwarenessCheckChaptersModel.FindOneByCheckAndChapter(l.ctx, check.CheckId, chapterID)
		if err != nil {
			return nil, fmt.Errorf("query new awareness check chapter: %w", err)
		}
	}

	chapterPoints := sources.pointsByChapter[chapterID]
	if len(chapterPoints) == 0 {
		return nil, fmt.Errorf("awareness check chapter has no points")
	}

	prevScoreByAwareness, err := loadPrevAwarenessScores(l.ctx, l.svcCtx, check, chapterID)
	if err != nil {
		return nil, err
	}
	for _, input := range req.Scores {
		point, ok := sources.pointByID[input.AwarenessId]
		if !ok || point.ChapterId != chapterID {
			return nil, fmt.Errorf("awareness point %d is not in chapter %d", input.AwarenessId, chapterID)
		}
		scored, err := scoreCheckPoint(input.SelfScore, checkHumanScore(point.ReferenceMin, point.ReferenceMax), point.BetterDirection)
		if err != nil {
			return nil, err
		}

		scoreRecord := &model.AwarenessCheckScore{
			CheckId:     check.CheckId,
			UserId:      userID,
			ChapterId:   chapterID,
			AwarenessId: point.AwarenessId,
			SelfScore:   scored.SelfScore,
			Score:       scored.Score,
			RefScore:    scored.RefScore,
			Delta:       scored.Delta,
		}
		if prevScore, ok := prevScoreByAwareness[point.AwarenessId]; ok {
			scoreRecord.PrevScore = sqlFloat(prevScore)
			scoreRecord.ScoreChange = sqlFloat(roundCheckScore(scored.Score - prevScore))
		}
		if _, err = l.svcCtx.AwarenessCheckScoresModel.Upsert(l.ctx, scoreRecord); err != nil {
			return nil, fmt.Errorf("save awareness check score: %w", err)
		}
	}

	savedScores, err := l.svcCtx.AwarenessCheckScoresModel.FindByCheckAndChapter(l.ctx, check.CheckId, chapterID)
	if err != nil {
		return nil, fmt.Errorf("query awareness check scores: %w", err)
	}
	chapter.ScoredPoints = int64(len(savedScores))
	chapter.Status = checkChapterStatusInProgress
	if err = l.svcCtx.AwarenessCheckChaptersModel.Update(l.ctx, chapter); err != nil {
		return nil, fmt.Errorf("update awareness check chapter: %w", err)
	}

	if check.Status == checkStatusDraft {
		check.Status = checkStatusInProgress
		if err = l.svcCtx.AwarenessChecksModel.Update(l.ctx, check); err != nil {
			return nil, fmt.Errorf("update awareness check: %w", err)
		}
	}
	check, err = l.svcCtx.AwarenessChecksModel.FindOne(l.ctx, check.CheckId)
	if err != nil {
		return nil, fmt.Errorf("reload awareness check: %w", err)
	}

	data, err := buildAwarenessCheckChapterData(l.ctx, l.svcCtx, check, chapterID)
	if err != nil {
		return nil, err
	}
	return &types.AwarenessCheckChapterResp{Code: 0, Message: "ok", Data: data}, nil
}

type SubmitAwarenessCheckChapterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSubmitAwarenessCheckChapterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitAwarenessCheckChapterLogic {
	return &SubmitAwarenessCheckChapterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SubmitAwarenessCheckChapterLogic) SubmitAwarenessCheckChapter(chapterID uint64) (*types.AwarenessCheckChapterResp, error) {
	if !hasAwarenessCheckStorage(l.svcCtx) {
		return &types.AwarenessCheckChapterResp{Code: 0, Message: "ok", Data: types.AwarenessCheckChapterData{}}, nil
	}

	userID := currentUserID(l.ctx)
	check, err := loadOrCreateCurrentAwarenessCheck(l.ctx, l.svcCtx, userID)
	if err != nil {
		return nil, err
	}
	chapter, err := l.svcCtx.AwarenessCheckChaptersModel.FindOneByCheckAndChapter(l.ctx, check.CheckId, chapterID)
	if err != nil {
		return nil, fmt.Errorf("query awareness check chapter: %w", err)
	}
	if chapter.UserId != userID {
		return nil, fmt.Errorf("awareness check chapter not found")
	}
	if chapter.Status == checkStatusCompleted {
		return nil, fmt.Errorf("chapter scores must be saved before submit")
	}

	sources, err := loadAwarenessCheckSources(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}
	chapterPoints := sources.pointsByChapter[chapterID]
	if len(chapterPoints) == 0 {
		return nil, fmt.Errorf("awareness check chapter has no points")
	}
	savedScores, err := l.svcCtx.AwarenessCheckScoresModel.FindByCheckAndChapter(l.ctx, check.CheckId, chapterID)
	if err != nil {
		return nil, fmt.Errorf("query awareness check scores: %w", err)
	}
	if len(savedScores) != len(chapterPoints) {
		return nil, fmt.Errorf("all chapter points must be saved before submit")
	}

	pointScores := make([]checkPointScore, 0, len(savedScores))
	for _, score := range savedScores {
		pointScores = append(pointScores, checkPointScore{
			SelfScore: score.SelfScore,
			Score:     score.Score,
			RefScore:  score.RefScore,
			Delta:     score.Delta,
		})
	}
	chapterScore, ok := avgCheckPoints(pointScores)
	if !ok {
		return nil, fmt.Errorf("chapter score unavailable")
	}

	prevScoreByChapter, err := loadPrevChapterScores(l.ctx, l.svcCtx, check)
	if err != nil {
		return nil, err
	}
	chapter.Score = sqlFloat(chapterScore.Score)
	chapter.RefScore = sqlFloat(chapterScore.RefScore)
	chapter.Delta = sqlFloat(checkDelta(chapterScore))
	if prevScore, ok := prevScoreByChapter[chapterID]; ok {
		chapter.PrevScore = sqlFloat(prevScore)
		chapter.ScoreChange = sqlFloat(roundCheckScore(chapterScore.Score - prevScore))
	}
	chapter.ScoredPoints = int64(len(savedScores))
	chapter.Status = checkStatusCompleted
	chapter.SubmittedAt = sql.NullTime{Time: time.Now(), Valid: true}
	if err = l.svcCtx.AwarenessCheckChaptersModel.Update(l.ctx, chapter); err != nil {
		return nil, fmt.Errorf("submit awareness check chapter: %w", err)
	}

	if err = refreshAwarenessCheckOverall(l.ctx, l.svcCtx, check.CheckId); err != nil {
		return nil, err
	}
	check, err = l.svcCtx.AwarenessChecksModel.FindOne(l.ctx, check.CheckId)
	if err != nil {
		return nil, fmt.Errorf("reload awareness check: %w", err)
	}
	check.Status = checkStatusCompleted
	if !check.CompletedAt.Valid {
		check.CompletedAt = sql.NullTime{Time: time.Now(), Valid: true}
	}
	if err = l.svcCtx.AwarenessChecksModel.Update(l.ctx, check); err != nil {
		return nil, fmt.Errorf("complete chapter attempt: %w", err)
	}
	check, err = l.svcCtx.AwarenessChecksModel.FindOne(l.ctx, check.CheckId)
	if err != nil {
		return nil, fmt.Errorf("reload completed awareness check: %w", err)
	}
	data, err := buildAwarenessCheckChapterData(l.ctx, l.svcCtx, check, chapterID)
	if err != nil {
		return nil, err
	}
	return &types.AwarenessCheckChapterResp{Code: 0, Message: "ok", Data: data}, nil
}

type ListAwarenessCheckHistoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListAwarenessCheckHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAwarenessCheckHistoryLogic {
	return &ListAwarenessCheckHistoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListAwarenessCheckHistoryLogic) ListAwarenessCheckHistory() (*types.AwarenessCheckHistoryResp, error) {
	if l.svcCtx.AwarenessChecksModel == nil {
		return &types.AwarenessCheckHistoryResp{Code: 0, Message: "ok", Data: types.AwarenessCheckHistoryData{}}, nil
	}
	items, err := l.svcCtx.AwarenessChecksModel.FindByUserID(l.ctx, currentUserID(l.ctx), 50)
	if err != nil {
		return nil, fmt.Errorf("query awareness check history: %w", err)
	}

	list := make([]types.AwarenessCheckInfo, 0, len(items))
	for i := range items {
		item := items[i]
		list = append(list, awarenessCheckToInfo(&item))
	}
	return &types.AwarenessCheckHistoryResp{
		Code:    0,
		Message: "ok",
		Data: types.AwarenessCheckHistoryData{
			List: list,
		},
	}, nil
}

type GetAwarenessCheckDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAwarenessCheckDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAwarenessCheckDetailLogic {
	return &GetAwarenessCheckDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAwarenessCheckDetailLogic) GetAwarenessCheckDetail(checkID uint64) (*types.AwarenessCheckDetailResp, error) {
	if !hasAwarenessCheckStorage(l.svcCtx) {
		return &types.AwarenessCheckDetailResp{Code: 0, Message: "ok", Data: types.AwarenessCheckDetailData{}}, nil
	}

	check, err := l.svcCtx.AwarenessChecksModel.FindOne(l.ctx, checkID)
	if err != nil {
		return nil, fmt.Errorf("query awareness check: %w", err)
	}
	if check.UserId != currentUserID(l.ctx) {
		return nil, fmt.Errorf("awareness check not found")
	}
	chapters, err := l.svcCtx.AwarenessCheckChaptersModel.FindByCheckID(l.ctx, check.CheckId)
	if err != nil {
		return nil, fmt.Errorf("query awareness check chapters: %w", err)
	}
	scores, err := l.svcCtx.AwarenessCheckScoresModel.FindByCheckID(l.ctx, check.CheckId)
	if err != nil {
		return nil, fmt.Errorf("query awareness check scores: %w", err)
	}
	sources, err := loadAwarenessCheckSources(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}

	chapterInfos := make([]types.AwarenessCheckChapterInfo, 0, len(chapters))
	for _, chapter := range sortCheckChapters(chapters, sources.chapterByID) {
		source := sources.chapterByID[chapter.ChapterId]
		chapterInfos = append(chapterInfos, awarenessCheckChapterToInfo(&chapter, &source))
	}
	scoreInfos := make([]types.AwarenessCheckScoreInfo, 0, len(scores))
	for _, score := range scores {
		point, ok := sources.pointByID[score.AwarenessId]
		if !ok {
			continue
		}
		scoreInfos = append(scoreInfos, awarenessCheckScoreToInfo(&score, &point))
	}
	compare, err := buildAwarenessCheckCompare(l.ctx, l.svcCtx, check, 0)
	if err != nil {
		return nil, err
	}
	return &types.AwarenessCheckDetailResp{
		Code:    0,
		Message: "ok",
		Data: types.AwarenessCheckDetailData{
			Check:    awarenessCheckToInfo(check),
			Chapters: chapterInfos,
			Scores:   scoreInfos,
			Compare:  compare,
		},
	}, nil
}

type GetAwarenessCheckTrendsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAwarenessCheckTrendsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAwarenessCheckTrendsLogic {
	return &GetAwarenessCheckTrendsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAwarenessCheckTrendsLogic) GetAwarenessCheckTrends() (*types.AwarenessCheckTrendsResp, error) {
	if !hasAwarenessCheckStorage(l.svcCtx) {
		return &types.AwarenessCheckTrendsResp{Code: 0, Message: "ok", Data: types.AwarenessCheckTrendsData{}}, nil
	}

	userID := currentUserID(l.ctx)
	checks, err := l.svcCtx.AwarenessChecksModel.FindByUserID(l.ctx, userID, 0)
	if err != nil {
		return nil, fmt.Errorf("query awareness check trends: %w", err)
	}
	sort.SliceStable(checks, func(i, j int) bool {
		if checks[i].StartedAt.Equal(checks[j].StartedAt) {
			return checks[i].CheckId < checks[j].CheckId
		}
		return checks[i].StartedAt.Before(checks[j].StartedAt)
	})

	sources, err := loadAwarenessCheckSources(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}
	overall := make([]types.AwarenessCheckTrendPoint, 0, len(checks))
	chapterTrends := make([]types.AwarenessCheckChapterTrendPoint, 0)
	refDeltas := make([]types.AwarenessCheckChapterTrendPoint, 0)

	checkByID := make(map[uint64]model.AwarenessCheck, len(checks))
	for _, check := range checks {
		checkByID[check.CheckId] = check
	}
	completedChapters, err := l.svcCtx.AwarenessCheckChaptersModel.FindCompletedByUserID(l.ctx, userID, 0)
	if err != nil {
		return nil, fmt.Errorf("query completed awareness check chapters: %w", err)
	}
	sort.SliceStable(completedChapters, func(i, j int) bool {
		left := completedChapters[i]
		right := completedChapters[j]
		if left.SubmittedAt.Valid && right.SubmittedAt.Valid && !left.SubmittedAt.Time.Equal(right.SubmittedAt.Time) {
			return left.SubmittedAt.Time.Before(right.SubmittedAt.Time)
		}
		if left.CheckId == right.CheckId {
			return left.CheckChapterId < right.CheckChapterId
		}
		return left.CheckId < right.CheckId
	})

	latestByChapter := make(map[uint64]model.AwarenessCheckChapter, len(sources.chapters))
	totalChapters := int64(len(sources.chapters))
	for _, chapter := range completedChapters {
		parent, ok := checkByID[chapter.CheckId]
		if !ok || parent.Status == checkStatusAbandoned {
			continue
		}
		if chapter.Status != checkStatusCompleted || !chapter.Score.Valid {
			continue
		}
		source, ok := sources.chapterByID[chapter.ChapterId]
		if !ok {
			continue
		}
		point := types.AwarenessCheckChapterTrendPoint{
			CheckId:          chapter.CheckId,
			ChapterId:        chapter.ChapterId,
			ChapterNo:        source.ChapterNo,
			ChapterTitle:     source.ChapterTitle,
			ChapterFullTitle: source.ChapterFullTitle,
			SubmittedAt:      formatNullTime(chapter.SubmittedAt),
			Score:            nullableFloat(chapter.Score),
			RefScore:         nullableFloat(chapter.RefScore),
			Delta:            nullableFloat(chapter.Delta),
			HasPrevScore:     chapter.PrevScore.Valid,
			ScoreChange:      nullableFloat(chapter.ScoreChange),
		}
		chapterTrends = append(chapterTrends, point)
		refDeltas = append(refDeltas, point)

		latestByChapter[chapter.ChapterId] = chapter
		if totalChapters > 0 && int64(len(latestByChapter)) == totalChapters {
			score, ok := aggregateCompletedChapterScores(latestByChapter)
			if !ok {
				continue
			}
			overall = append(overall, types.AwarenessCheckTrendPoint{
				CheckId:       chapter.CheckId,
				StartedAt:     parent.StartedAt.Format("2006-01-02 15:04:05"),
				CompletedAt:   formatNullTime(chapter.SubmittedAt),
				DoneChapters:  totalChapters,
				TotalChapters: totalChapters,
				Score:         score.Score,
				RefScore:      score.RefScore,
				Delta:         checkDelta(score),
			})
		}
	}

	return &types.AwarenessCheckTrendsResp{
		Code:    0,
		Message: "ok",
		Data: types.AwarenessCheckTrendsData{
			Overall:         overall,
			Chapters:        chapterTrends,
			ReferenceDeltas: refDeltas,
		},
	}, nil
}

func hasAwarenessCheckStorage(svcCtx *svc.ServiceContext) bool {
	return svcCtx != nil &&
		svcCtx.AwarenessChecksModel != nil &&
		svcCtx.AwarenessCheckChaptersModel != nil &&
		svcCtx.AwarenessCheckScoresModel != nil &&
		svcCtx.AwarenessModel != nil &&
		svcCtx.ChaptersModel != nil
}

func loadOrCreateCurrentAwarenessCheck(ctx context.Context, svcCtx *svc.ServiceContext, userID uint64) (*model.AwarenessCheck, error) {
	sources, err := loadAwarenessCheckSources(ctx, svcCtx)
	if err != nil {
		return nil, err
	}

	current, err := svcCtx.AwarenessChecksModel.FindCurrentByUserID(ctx, userID)
	if err == nil {
		if err = ensureAwarenessCheckChapters(ctx, svcCtx, current, sources); err != nil {
			return nil, err
		}
		return svcCtx.AwarenessChecksModel.FindOne(ctx, current.CheckId)
	}
	if err != model.ErrNotFound {
		return nil, fmt.Errorf("query current awareness check: %w", err)
	}

	return createNewAwarenessCheck(ctx, svcCtx, userID, sources)
}

func createNewAwarenessCheck(ctx context.Context, svcCtx *svc.ServiceContext, userID uint64, sources awarenessCheckSources) (*model.AwarenessCheck, error) {
	prevCheckID := sql.NullInt64{}
	prev, prevErr := svcCtx.AwarenessChecksModel.FindLatestScoredByUserID(ctx, userID)
	if prevErr == nil {
		prevCheckID = sql.NullInt64{Int64: int64(prev.CheckId), Valid: true}
	} else if prevErr != model.ErrNotFound {
		return nil, fmt.Errorf("query previous awareness check: %w", prevErr)
	}

	now := time.Now()
	result, err := svcCtx.AwarenessChecksModel.Insert(ctx, &model.AwarenessCheck{
		UserId:        userID,
		Status:        checkStatusDraft,
		DoneChapters:  0,
		TotalChapters: int64(len(sources.chapters)),
		PrevCheckId:   prevCheckID,
		StartedAt:     now,
	})
	if err != nil {
		if isDuplicateEntryError(err) {
			current, currentErr := svcCtx.AwarenessChecksModel.FindCurrentByUserID(ctx, userID)
			if currentErr == nil {
				if ensureErr := ensureAwarenessCheckChapters(ctx, svcCtx, current, sources); ensureErr != nil {
					return nil, ensureErr
				}
				return svcCtx.AwarenessChecksModel.FindOne(ctx, current.CheckId)
			}
		}
		return nil, fmt.Errorf("create awareness check: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("read awareness check id: %w", err)
	}
	current, err := svcCtx.AwarenessChecksModel.FindOne(ctx, uint64(id))
	if err != nil {
		return nil, fmt.Errorf("load awareness check: %w", err)
	}
	if err = ensureAwarenessCheckChapters(ctx, svcCtx, current, sources); err != nil {
		return nil, err
	}
	return svcCtx.AwarenessChecksModel.FindOne(ctx, current.CheckId)
}

func startNewChapterAttempt(ctx context.Context, svcCtx *svc.ServiceContext, check *model.AwarenessCheck, sources awarenessCheckSources) (*model.AwarenessCheck, error) {
	if check == nil {
		return nil, fmt.Errorf("awareness check not found")
	}
	userID := check.UserId
	if check != nil && (check.Status == checkStatusDraft || check.Status == checkStatusInProgress) {
		check.Status = checkStatusCompleted
		if !check.CompletedAt.Valid {
			check.CompletedAt = sql.NullTime{Time: time.Now(), Valid: true}
		}
		if err := svcCtx.AwarenessChecksModel.Update(ctx, check); err != nil {
			return nil, fmt.Errorf("close previous chapter attempt: %w", err)
		}
	}
	return createNewAwarenessCheck(ctx, svcCtx, userID, sources)
}

func abandonCurrentAwarenessChecks(ctx context.Context, svcCtx *svc.ServiceContext, userID uint64) error {
	for {
		current, err := svcCtx.AwarenessChecksModel.FindCurrentByUserID(ctx, userID)
		if err == model.ErrNotFound {
			return nil
		}
		if err != nil {
			return fmt.Errorf("query current awareness check: %w", err)
		}
		current.Status = checkStatusAbandoned
		if err = svcCtx.AwarenessChecksModel.Update(ctx, current); err != nil {
			return fmt.Errorf("abandon current awareness check: %w", err)
		}
	}
}

func isDuplicateEntryError(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}

func ensureAwarenessCheckChapters(ctx context.Context, svcCtx *svc.ServiceContext, check *model.AwarenessCheck, sources awarenessCheckSources) error {
	existing, err := svcCtx.AwarenessCheckChaptersModel.FindByCheckID(ctx, check.CheckId)
	if err != nil {
		return fmt.Errorf("query awareness check chapters: %w", err)
	}
	existingByChapter := make(map[uint64]bool, len(existing))
	for _, item := range existing {
		existingByChapter[item.ChapterId] = true
	}

	prevScoreByChapter, err := loadPrevChapterScores(ctx, svcCtx, check)
	if err != nil {
		return err
	}
	for _, chapter := range sources.chapters {
		if existingByChapter[chapter.ChapterId] {
			continue
		}
		data := &model.AwarenessCheckChapter{
			CheckId:      check.CheckId,
			UserId:       check.UserId,
			ChapterId:    chapter.ChapterId,
			TotalPoints:  int64(len(sources.pointsByChapter[chapter.ChapterId])),
			ScoredPoints: 0,
			Status:       checkChapterStatusNotStarted,
		}
		if prevScore, ok := prevScoreByChapter[chapter.ChapterId]; ok {
			data.PrevScore = sqlFloat(prevScore)
		}
		if _, err = svcCtx.AwarenessCheckChaptersModel.Insert(ctx, data); err != nil {
			if isDuplicateEntryError(err) {
				continue
			}
			return fmt.Errorf("create awareness check chapter: %w", err)
		}
	}
	return nil
}

func latestCompletedChaptersByID(ctx context.Context, svcCtx *svc.ServiceContext, userID uint64, excludeCheckID uint64) (map[uint64]model.AwarenessCheckChapter, []model.AwarenessCheckChapter, error) {
	chapters, err := svcCtx.AwarenessCheckChaptersModel.FindCompletedByUserID(ctx, userID, 0)
	if err != nil {
		return nil, nil, fmt.Errorf("query completed awareness check chapters: %w", err)
	}

	latest := make(map[uint64]model.AwarenessCheckChapter)
	ordered := make([]model.AwarenessCheckChapter, 0)
	for _, chapter := range chapters {
		if chapter.CheckId == excludeCheckID {
			continue
		}
		if _, ok := latest[chapter.ChapterId]; ok {
			continue
		}
		latest[chapter.ChapterId] = chapter
		ordered = append(ordered, chapter)
	}
	return latest, ordered, nil
}

func buildAwarenessCheckCurrentData(ctx context.Context, svcCtx *svc.ServiceContext, check *model.AwarenessCheck) (types.AwarenessCheckCurrentData, error) {
	sources, err := loadAwarenessCheckSources(ctx, svcCtx)
	if err != nil {
		return types.AwarenessCheckCurrentData{}, err
	}
	if err = ensureAwarenessCheckChapters(ctx, svcCtx, check, sources); err != nil {
		return types.AwarenessCheckCurrentData{}, err
	}
	currentChapters, err := svcCtx.AwarenessCheckChaptersModel.FindByCheckID(ctx, check.CheckId)
	if err != nil {
		return types.AwarenessCheckCurrentData{}, fmt.Errorf("query awareness check chapters: %w", err)
	}
	latestCompleted, _, err := latestCompletedChaptersByID(ctx, svcCtx, check.UserId, check.CheckId)
	if err != nil {
		return types.AwarenessCheckCurrentData{}, err
	}
	currentByChapter := make(map[uint64]model.AwarenessCheckChapter, len(currentChapters))
	for _, chapter := range currentChapters {
		currentByChapter[chapter.ChapterId] = chapter
	}

	chapterInfos := make([]types.AwarenessCheckChapterInfo, 0, len(sources.chapters))
	missingChapters := make([]types.AwarenessCheckChapterInfo, 0)
	displayChapters := make([]model.AwarenessCheckChapter, 0, len(sources.chapters))
	for _, source := range sources.chapters {
		chapter, ok := currentByChapter[source.ChapterId]
		if latest, hasLatest := latestCompleted[source.ChapterId]; hasLatest && (!ok || chapter.Status == checkChapterStatusNotStarted) {
			chapter = latest
		}
		displayChapters = append(displayChapters, chapter)
	}
	for _, chapter := range sortCheckChapters(displayChapters, sources.chapterByID) {
		source := sources.chapterByID[chapter.ChapterId]
		info := awarenessCheckChapterToInfo(&chapter, &source)
		chapterInfos = append(chapterInfos, info)
		if _, ok := latestCompleted[chapter.ChapterId]; !ok && chapter.Status != checkStatusCompleted {
			missingChapters = append(missingChapters, info)
		}
	}
	compare, err := buildAwarenessCheckCompare(ctx, svcCtx, check, 0)
	if err != nil {
		return types.AwarenessCheckCurrentData{}, err
	}
	return types.AwarenessCheckCurrentData{
		Check:           awarenessCheckToInfo(aggregateCurrentCheck(check, displayChapters, int64(len(sources.chapters)))),
		Chapters:        chapterInfos,
		MissingChapters: missingChapters,
		Compare:         compare,
	}, nil
}

func aggregateCurrentCheck(check *model.AwarenessCheck, chapters []model.AwarenessCheckChapter, totalChapters int64) *model.AwarenessCheck {
	if check == nil {
		return nil
	}
	result := *check
	result.TotalChapters = totalChapters

	doneCount := int64(0)
	scores := make([]checkChapterScore, 0, len(chapters))
	var latestSubmittedAt sql.NullTime
	for _, chapter := range chapters {
		if chapter.Status != checkStatusCompleted || !chapter.Score.Valid || !chapter.RefScore.Valid {
			continue
		}
		doneCount++
		scores = append(scores, checkChapterScore{
			Status:   checkStatusCompleted,
			Score:    chapter.Score.Float64,
			RefScore: chapter.RefScore.Float64,
		})
		if chapter.SubmittedAt.Valid && (!latestSubmittedAt.Valid || chapter.SubmittedAt.Time.After(latestSubmittedAt.Time)) {
			latestSubmittedAt = chapter.SubmittedAt
		}
	}

	result.DoneChapters = doneCount
	result.Score = sql.NullFloat64{}
	result.RefScore = sql.NullFloat64{}
	result.Delta = sql.NullFloat64{}
	result.CompletedAt = sql.NullTime{}
	if doneCount == totalChapters && totalChapters > 0 {
		if score, ok := avgDoneChapters(scores); ok {
			result.Score = sqlFloat(score.Score)
			result.RefScore = sqlFloat(score.RefScore)
			result.Delta = sqlFloat(checkDelta(score))
		}
		result.Status = checkStatusCompleted
		result.CompletedAt = latestSubmittedAt
	} else if doneCount > 0 && result.Status == checkStatusDraft {
		result.Status = checkStatusInProgress
	}

	return &result
}

func aggregateCompletedChapterScores(chapters map[uint64]model.AwarenessCheckChapter) (checkChapterScore, bool) {
	scores := make([]checkChapterScore, 0, len(chapters))
	for _, chapter := range chapters {
		if chapter.Status != checkStatusCompleted || !chapter.Score.Valid || !chapter.RefScore.Valid {
			return checkChapterScore{}, false
		}
		scores = append(scores, checkChapterScore{
			Status:   checkStatusCompleted,
			Score:    chapter.Score.Float64,
			RefScore: chapter.RefScore.Float64,
		})
	}
	return avgDoneChapters(scores)
}

func buildAwarenessCheckChapterData(ctx context.Context, svcCtx *svc.ServiceContext, check *model.AwarenessCheck, chapterID uint64) (types.AwarenessCheckChapterData, error) {
	sources, err := loadAwarenessCheckSources(ctx, svcCtx)
	if err != nil {
		return types.AwarenessCheckChapterData{}, err
	}
	source, ok := sources.chapterByID[chapterID]
	if !ok {
		return types.AwarenessCheckChapterData{}, fmt.Errorf("chapter not found")
	}
	if err = ensureAwarenessCheckChapters(ctx, svcCtx, check, sources); err != nil {
		return types.AwarenessCheckChapterData{}, err
	}
	chapter, err := svcCtx.AwarenessCheckChaptersModel.FindOneByCheckAndChapter(ctx, check.CheckId, chapterID)
	if err != nil {
		return types.AwarenessCheckChapterData{}, fmt.Errorf("query awareness check chapter: %w", err)
	}
	savedScores, err := svcCtx.AwarenessCheckScoresModel.FindByCheckAndChapter(ctx, check.CheckId, chapterID)
	if err != nil {
		return types.AwarenessCheckChapterData{}, fmt.Errorf("query awareness check scores: %w", err)
	}
	savedByAwareness := make(map[uint64]model.AwarenessCheckScore, len(savedScores))
	for _, score := range savedScores {
		savedByAwareness[score.AwarenessId] = score
	}
	prevScoreByAwareness, err := loadPrevAwarenessScores(ctx, svcCtx, check, chapterID)
	if err != nil {
		return types.AwarenessCheckChapterData{}, err
	}

	points := make([]types.AwarenessCheckPointInfo, 0, len(sources.pointsByChapter[chapterID]))
	for _, point := range sources.pointsByChapter[chapterID] {
		humanScore := checkHumanScore(point.ReferenceMin, point.ReferenceMax)
		scored, err := scoreCheckPoint(50, humanScore, point.BetterDirection)
		if err != nil {
			return types.AwarenessCheckChapterData{}, err
		}
		info := awarenessCheckPointToInfo(point, scored)
		if saved, ok := savedByAwareness[point.AwarenessId]; ok {
			info.SelfScore = saved.SelfScore
			info.Score = saved.Score
			info.RefScore = saved.RefScore
			info.Delta = saved.Delta
			if saved.PrevScore.Valid {
				info.PrevScore = saved.PrevScore.Float64
				info.HasPrevScore = true
			}
			if saved.ScoreChange.Valid {
				info.ScoreChange = saved.ScoreChange.Float64
			}
		} else if prevScore, ok := prevScoreByAwareness[point.AwarenessId]; ok {
			info.PrevScore = prevScore
			info.HasPrevScore = true
			info.ScoreChange = roundCheckScore(info.Score - prevScore)
		}
		points = append(points, info)
	}
	compare, err := buildAwarenessCheckCompare(ctx, svcCtx, check, chapterID)
	if err != nil {
		return types.AwarenessCheckChapterData{}, err
	}
	return types.AwarenessCheckChapterData{
		Check:   awarenessCheckToInfo(check),
		Chapter: awarenessCheckChapterToInfo(chapter, &source),
		Points:  points,
		Compare: compare,
	}, nil
}

func refreshAwarenessCheckOverall(ctx context.Context, svcCtx *svc.ServiceContext, checkID uint64) error {
	check, err := svcCtx.AwarenessChecksModel.FindOne(ctx, checkID)
	if err != nil {
		return fmt.Errorf("query awareness check: %w", err)
	}
	chapters, err := svcCtx.AwarenessCheckChaptersModel.FindByCheckID(ctx, checkID)
	if err != nil {
		return fmt.Errorf("query awareness check chapters: %w", err)
	}

	doneCount := int64(0)
	scores := make([]checkChapterScore, 0, len(chapters))
	for _, chapter := range chapters {
		if chapter.Status != checkStatusCompleted || !chapter.Score.Valid || !chapter.RefScore.Valid {
			scores = append(scores, checkChapterScore{Status: chapter.Status})
			continue
		}
		doneCount++
		scores = append(scores, checkChapterScore{
			Status:   checkStatusCompleted,
			Score:    chapter.Score.Float64,
			RefScore: chapter.RefScore.Float64,
		})
	}

	check.DoneChapters = doneCount
	if score, ok := avgDoneChapters(scores); ok {
		check.Score = sqlFloat(score.Score)
		check.RefScore = sqlFloat(score.RefScore)
		check.Delta = sqlFloat(checkDelta(score))
	}
	if doneCount > 0 {
		check.Status = checkStatusInProgress
	}
	if doneCount == check.TotalChapters && check.TotalChapters > 0 {
		check.Status = checkStatusCompleted
		if !check.CompletedAt.Valid {
			check.CompletedAt = sql.NullTime{Time: time.Now(), Valid: true}
		}
	}
	if err = svcCtx.AwarenessChecksModel.Update(ctx, check); err != nil {
		return fmt.Errorf("update awareness check: %w", err)
	}
	return nil
}

func buildAwarenessCheckCompare(ctx context.Context, svcCtx *svc.ServiceContext, check *model.AwarenessCheck, chapterFilter uint64) (types.AwarenessCheckCompareInfo, error) {
	compare := types.AwarenessCheckCompareInfo{
		ImprovedChapters: []types.AwarenessCheckChangeInfo{},
		DeclinedChapters: []types.AwarenessCheckChangeInfo{},
		ImprovedPoints:   []types.AwarenessCheckChangeInfo{},
		DeclinedPoints:   []types.AwarenessCheckChangeInfo{},
	}
	if check.PrevCheckId.Valid {
		compare.PrevCheckId = uint64(check.PrevCheckId.Int64)
		prev, err := svcCtx.AwarenessChecksModel.FindOne(ctx, uint64(check.PrevCheckId.Int64))
		if err == nil && check.Score.Valid && prev.Score.Valid {
			compare.ScoreChange = roundCheckScore(check.Score.Float64 - prev.Score.Float64)
		} else if err != nil && err != model.ErrNotFound {
			return compare, fmt.Errorf("query previous awareness check: %w", err)
		}
	}

	sources, err := loadAwarenessCheckSources(ctx, svcCtx)
	if err != nil {
		return compare, err
	}
	chapters, err := svcCtx.AwarenessCheckChaptersModel.FindByCheckID(ctx, check.CheckId)
	if err != nil {
		return compare, fmt.Errorf("query awareness check chapters: %w", err)
	}
	chapterChanges := make([]types.AwarenessCheckChangeInfo, 0)
	for _, chapter := range chapters {
		source := sources.chapterByID[chapter.ChapterId]
		if chapter.Status != checkStatusCompleted {
			if source.ChapterFullTitle != "" {
				compare.MissingChapters = append(compare.MissingChapters, source.ChapterFullTitle)
			}
			continue
		}
		if chapter.ScoreChange.Valid {
			compare.SameChapterCount++
			chapterChanges = append(chapterChanges, types.AwarenessCheckChangeInfo{
				Scope:       "chapter",
				ChapterId:   chapter.ChapterId,
				Title:       source.ChapterFullTitle,
				ScoreChange: chapter.ScoreChange.Float64,
			})
		}
	}

	scores, err := svcCtx.AwarenessCheckScoresModel.FindByCheckID(ctx, check.CheckId)
	if err != nil {
		return compare, fmt.Errorf("query awareness check scores: %w", err)
	}
	pointChanges := make([]types.AwarenessCheckChangeInfo, 0)
	for _, score := range scores {
		if chapterFilter > 0 && score.ChapterId != chapterFilter {
			continue
		}
		if !score.ScoreChange.Valid {
			continue
		}
		point := sources.pointByID[score.AwarenessId]
		pointChanges = append(pointChanges, types.AwarenessCheckChangeInfo{
			Scope:       "point",
			ChapterId:   score.ChapterId,
			AwarenessId: score.AwarenessId,
			Title:       point.PointTitle,
			ScoreChange: score.ScoreChange.Float64,
		})
	}

	compare.ImprovedChapters = topPositiveChanges(chapterChanges, 3)
	compare.DeclinedChapters = topNegativeChanges(chapterChanges, 3)
	compare.ImprovedPoints = topPositiveChanges(pointChanges, 3)
	compare.DeclinedPoints = topNegativeChanges(pointChanges, 3)
	return compare, nil
}

func loadAwarenessCheckSources(ctx context.Context, svcCtx *svc.ServiceContext) (awarenessCheckSources, error) {
	chapters, err := svcCtx.ChaptersModel.FindAll(ctx)
	if err != nil {
		return awarenessCheckSources{}, fmt.Errorf("query chapters: %w", err)
	}
	points, err := svcCtx.AwarenessModel.FindEligible(ctx)
	if err != nil {
		return awarenessCheckSources{}, fmt.Errorf("query awareness points: %w", err)
	}

	source := awarenessCheckSources{
		chapters:        chapters,
		points:          points,
		chapterByID:     make(map[uint64]model.Chapter, len(chapters)),
		pointsByChapter: make(map[uint64][]model.Awareness),
		pointByID:       make(map[uint64]model.Awareness, len(points)),
	}
	for _, chapter := range chapters {
		source.chapterByID[chapter.ChapterId] = chapter
	}
	for _, point := range points {
		source.pointsByChapter[point.ChapterId] = append(source.pointsByChapter[point.ChapterId], point)
		source.pointByID[point.AwarenessId] = point
	}
	for chapterID := range source.pointsByChapter {
		sort.SliceStable(source.pointsByChapter[chapterID], func(i, j int) bool {
			left := source.pointsByChapter[chapterID][i]
			right := source.pointsByChapter[chapterID][j]
			if left.SortOrderGlobal == right.SortOrderGlobal {
				return left.AwarenessId < right.AwarenessId
			}
			return left.SortOrderGlobal < right.SortOrderGlobal
		})
	}
	return source, nil
}

func loadPrevChapterScores(ctx context.Context, svcCtx *svc.ServiceContext, check *model.AwarenessCheck) (map[uint64]float64, error) {
	result := map[uint64]float64{}
	if check == nil {
		return result, nil
	}
	latest, _, err := latestCompletedChaptersByID(ctx, svcCtx, check.UserId, check.CheckId)
	if err != nil {
		return nil, err
	}
	for chapterID, chapter := range latest {
		if chapter.Score.Valid {
			result[chapterID] = chapter.Score.Float64
		}
	}
	return result, nil
}

func loadPrevAwarenessScores(ctx context.Context, svcCtx *svc.ServiceContext, check *model.AwarenessCheck, chapterID uint64) (map[uint64]float64, error) {
	result := map[uint64]float64{}
	if check == nil {
		return result, nil
	}
	latest, _, err := latestCompletedChaptersByID(ctx, svcCtx, check.UserId, check.CheckId)
	if err != nil {
		return nil, err
	}
	prevChapter, ok := latest[chapterID]
	if !ok {
		return result, nil
	}
	scores, err := svcCtx.AwarenessCheckScoresModel.FindByCheckAndChapter(ctx, prevChapter.CheckId, chapterID)
	if err != nil {
		return nil, fmt.Errorf("query previous awareness check scores: %w", err)
	}
	for _, score := range scores {
		result[score.AwarenessId] = score.Score
	}
	return result, nil
}

func awarenessCheckToInfo(item *model.AwarenessCheck) types.AwarenessCheckInfo {
	if item == nil {
		return types.AwarenessCheckInfo{}
	}
	return types.AwarenessCheckInfo{
		CheckId:       item.CheckId,
		Status:        item.Status,
		DoneChapters:  item.DoneChapters,
		TotalChapters: item.TotalChapters,
		Score:         nullableFloat(item.Score),
		RefScore:      nullableFloat(item.RefScore),
		Delta:         nullableFloat(item.Delta),
		PrevCheckId:   uint64(nullableInt64(item.PrevCheckId)),
		StartedAt:     item.StartedAt.Format("2006-01-02 15:04:05"),
		CompletedAt:   formatNullTime(item.CompletedAt),
		CreatedAt:     item.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     item.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func awarenessCheckChapterToInfo(item *model.AwarenessCheckChapter, chapter *model.Chapter) types.AwarenessCheckChapterInfo {
	if item == nil {
		return types.AwarenessCheckChapterInfo{}
	}
	info := types.AwarenessCheckChapterInfo{
		CheckChapterId: item.CheckChapterId,
		CheckId:        item.CheckId,
		ChapterId:      item.ChapterId,
		TotalPoints:    item.TotalPoints,
		ScoredPoints:   item.ScoredPoints,
		Score:          nullableFloat(item.Score),
		RefScore:       nullableFloat(item.RefScore),
		Delta:          nullableFloat(item.Delta),
		PrevScore:      nullableFloat(item.PrevScore),
		HasPrevScore:   item.PrevScore.Valid,
		ScoreChange:    nullableFloat(item.ScoreChange),
		Status:         item.Status,
		SubmittedAt:    formatNullTime(item.SubmittedAt),
	}
	if chapter != nil {
		info.ChapterNo = chapter.ChapterNo
		info.ChapterTitle = chapter.ChapterTitle
		info.ChapterFullTitle = chapter.ChapterFullTitle
	}
	return info
}

func awarenessCheckPointToInfo(point model.Awareness, scored checkPointScore) types.AwarenessCheckPointInfo {
	direction := cleanCheckDirection(point.BetterDirection)
	return types.AwarenessCheckPointInfo{
		AwarenessId:   point.AwarenessId,
		ChapterId:     point.ChapterId,
		SectionId:     point.SectionId,
		Title:         point.PointTitle,
		Summary:       nullableString(point.Summary),
		Details:       nullableString(point.Details),
		OrderNo:       point.SortOrderGlobal,
		HumanScore:    scored.HumanScore,
		Direction:     direction,
		DirectionText: checkDirectionText(direction),
		SelfScore:     scored.SelfScore,
		Score:         scored.Score,
		RefScore:      scored.RefScore,
		Delta:         scored.Delta,
	}
}

func awarenessCheckScoreToInfo(item *model.AwarenessCheckScore, point *model.Awareness) types.AwarenessCheckScoreInfo {
	if item == nil {
		return types.AwarenessCheckScoreInfo{}
	}
	direction := ""
	humanScore := 50.0
	title := ""
	summary := ""
	if point != nil {
		direction = cleanCheckDirection(point.BetterDirection)
		humanScore = checkHumanScore(point.ReferenceMin, point.ReferenceMax)
		title = point.PointTitle
		summary = nullableString(point.Summary)
	}
	return types.AwarenessCheckScoreInfo{
		ScoreId:       item.ScoreId,
		CheckId:       item.CheckId,
		ChapterId:     item.ChapterId,
		AwarenessId:   item.AwarenessId,
		Title:         title,
		Summary:       summary,
		SelfScore:     item.SelfScore,
		HumanScore:    humanScore,
		Direction:     direction,
		DirectionText: checkDirectionText(direction),
		Score:         item.Score,
		RefScore:      item.RefScore,
		Delta:         item.Delta,
		PrevScore:     nullableFloat(item.PrevScore),
		HasPrevScore:  item.PrevScore.Valid,
		ScoreChange:   nullableFloat(item.ScoreChange),
		CreatedAt:     item.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     item.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func sortCheckChapters(items []model.AwarenessCheckChapter, chapters map[uint64]model.Chapter) []model.AwarenessCheckChapter {
	result := append([]model.AwarenessCheckChapter(nil), items...)
	sort.SliceStable(result, func(i, j int) bool {
		left := chapters[result[i].ChapterId]
		right := chapters[result[j].ChapterId]
		if left.SortOrder == right.SortOrder {
			return result[i].ChapterId < result[j].ChapterId
		}
		return left.SortOrder < right.SortOrder
	})
	return result
}

func topPositiveChanges(items []types.AwarenessCheckChangeInfo, limit int) []types.AwarenessCheckChangeInfo {
	result := make([]types.AwarenessCheckChangeInfo, 0)
	for _, item := range items {
		if item.ScoreChange > 0 {
			result = append(result, item)
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].ScoreChange > result[j].ScoreChange
	})
	return limitChanges(result, limit)
}

func topNegativeChanges(items []types.AwarenessCheckChangeInfo, limit int) []types.AwarenessCheckChangeInfo {
	result := make([]types.AwarenessCheckChangeInfo, 0)
	for _, item := range items {
		if item.ScoreChange < 0 {
			result = append(result, item)
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].ScoreChange < result[j].ScoreChange
	})
	return limitChanges(result, limit)
}

func limitChanges(items []types.AwarenessCheckChangeInfo, limit int) []types.AwarenessCheckChangeInfo {
	if limit <= 0 || len(items) <= limit {
		return items
	}
	return items[:limit]
}

func checkDirectionText(direction string) string {
	if cleanCheckDirection(direction) == checkDirectionLower {
		return "越低越好"
	}
	return "越高越好"
}

func sqlFloat(value float64) sql.NullFloat64 {
	return sql.NullFloat64{Float64: roundCheckScore(value), Valid: true}
}

func nullableFloat(value sql.NullFloat64) float64 {
	if !value.Valid {
		return 0
	}
	return roundCheckScore(value.Float64)
}

func formatNullTime(value sql.NullTime) string {
	if !value.Valid {
		return ""
	}
	return value.Time.Format("2006-01-02 15:04:05")
}
