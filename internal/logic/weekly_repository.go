package logic

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"api/internal/svc"
	"api/internal/types"
	"api/model"
)

type weeklyVoteBundle struct {
	Vote          *model.WeeklyTopicVotes
	Candidates    []types.VoteCandidateInfo
	HasVoted      bool
	TodayHasVoted bool
	UserChoice    uint64
	TodayChoice   uint64
	TodayVotedAt  string
	MyRecords     []types.VoteRecordInfo
	Winner        *types.VoteCandidateInfo
}

func loadOrCreateCurrentWeeklyVote(ctx context.Context, svcCtx *svc.ServiceContext, now time.Time, userID uint64) (*weeklyVoteBundle, error) {
	weekStart := currentWeekStart(now)
	vote, err := ensureWeeklyVote(ctx, svcCtx, weekStart)
	if err != nil {
		return nil, err
	}

	candidates, err := listWeeklyVoteCandidates(ctx, svcCtx, vote.Id, vote.WeekStartDate)
	if err != nil {
		return nil, err
	}

	bundle := &weeklyVoteBundle{
		Vote:       vote,
		Candidates: candidates,
	}

	myRecords, recordErr := listWeeklyVoteRecordsByUser(ctx, svcCtx, vote.Id, userID)
	if recordErr != nil {
		return nil, recordErr
	}
	bundle.MyRecords = myRecords
	if len(myRecords) > 0 {
		bundle.HasVoted = true
		bundle.UserChoice = myRecords[0].CandidateId
	}
	if todayRecord, ok := findTodayVoteRecord(myRecords, now); ok {
		bundle.TodayHasVoted = true
		bundle.TodayChoice = todayRecord.CandidateId
		bundle.TodayVotedAt = todayRecord.CreatedAt
	}

	if winner, ok := pickWinningCandidate(candidates); ok {
		bundle.Winner = &winner
		if !vote.ResultTopicId.Valid || uint64(vote.ResultTopicId.Int64) != winner.TopicId {
			vote.ResultTopicId = sql.NullInt64{Int64: int64(winner.TopicId), Valid: true}
			if updateErr := svcCtx.WeeklyTopicVotesModel.Update(ctx, vote); updateErr != nil {
				return nil, updateErr
			}
		}
	}

	if now.After(vote.VoteEndAt) && vote.Status != "closed" {
		vote.Status = "closed"
		if updateErr := svcCtx.WeeklyTopicVotesModel.Update(ctx, vote); updateErr != nil {
			return nil, updateErr
		}
	}

	return bundle, nil
}

func ensureWeeklyVote(ctx context.Context, svcCtx *svc.ServiceContext, weekStart time.Time) (*model.WeeklyTopicVotes, error) {
	vote, err := svcCtx.WeeklyTopicVotesModel.FindOneByWeekStartDate(ctx, weekStart)
	if err == nil {
		return vote, nil
	}
	if err != model.ErrNotFound {
		return nil, err
	}

	if err := seedWeeklyVote(ctx, svcCtx, weekStart); err != nil {
		return nil, err
	}

	return svcCtx.WeeklyTopicVotesModel.FindOneByWeekStartDate(ctx, weekStart)
}

func seedWeeklyVote(ctx context.Context, svcCtx *svc.ServiceContext, weekStart time.Time) error {
	var topics []model.Topics
	candidateStart, candidateEnd := voteCandidateWindow(weekStart)
	query := "select id, title, summary, description, order_no, status, schedule_date, created_at, updated_at from topics where status = 1 and schedule_date >= ? and schedule_date <= ? order by schedule_date asc, order_no asc, id asc limit 7"
	if err := svcCtx.DB.QueryRowsCtx(ctx, &topics, query, candidateStart, candidateEnd); err != nil {
		return fmt.Errorf("query weekly vote topics: %w", err)
	}
	if len(topics) == 0 {
		return fmt.Errorf("no scheduled topics available for weekly vote window")
	}

	vote := &model.WeeklyTopicVotes{
		WeekStartDate: weekStart,
		VoteStartAt:   weekStart,
		VoteEndAt:     defaultVoteEndTime(weekStart),
		Status:        "active",
		ResultTopicId: sql.NullInt64{},
	}

	result, err := svcCtx.WeeklyTopicVotesModel.Insert(ctx, vote)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil
		}
		return err
	}

	voteID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	for index, topic := range topics {
		_, err = svcCtx.WeeklyTopicVoteCandidatesModel.Insert(ctx, &model.WeeklyTopicVoteCandidates{
			WeeklyVoteId: uint64(voteID),
			TopicId:      topic.Id,
			TopicTitle:   topic.Title,
			TopicSummary: topic.Summary,
			SortNo:       int64(index + 1),
		})
		if err != nil && !strings.Contains(err.Error(), "Duplicate entry") {
			return err
		}
	}

	return nil
}

func listWeeklyVoteCandidates(ctx context.Context, svcCtx *svc.ServiceContext, weeklyVoteID uint64, weekStart time.Time) ([]types.VoteCandidateInfo, error) {
	type voteCandidateRow struct {
		Id           uint64 `db:"id"`
		TopicId      uint64 `db:"topic_id"`
		TopicTitle   string `db:"topic_title"`
		TopicSummary string `db:"topic_summary"`
		SortNo       int64  `db:"sort_no"`
		VoteCount    int64  `db:"vote_count"`
	}

	var rows []voteCandidateRow
	query := `
		select c.id, c.topic_id, c.topic_title, c.topic_summary, c.sort_no, count(r.id) as vote_count
		from weekly_topic_vote_candidates c
		left join weekly_topic_vote_records r on r.candidate_id = c.id
		where c.weekly_vote_id = ?
		group by c.id, c.topic_id, c.topic_title, c.topic_summary, c.sort_no
		order by vote_count desc, c.sort_no asc, c.id asc`
	if err := svcCtx.DB.QueryRowsCtx(ctx, &rows, query, weeklyVoteID); err != nil {
		return nil, err
	}

	items := make([]types.VoteCandidateInfo, 0, len(rows))
	for _, row := range rows {
		candidateDate := voteCandidateDate(weekStart, row.SortNo)
		items = append(items, types.VoteCandidateInfo{
			Id:             row.Id,
			TopicId:        row.TopicId,
			TopicTitle:     row.TopicTitle,
			TopicSummary:   row.TopicSummary,
			TopicDate:      candidateDate.Format("2006-01-02"),
			TopicDateLabel: voteCandidateDateLabel(candidateDate),
			VoteCount:      row.VoteCount,
			SortNo:         row.SortNo,
		})
	}

	return items, nil
}

func listWeeklyVoteRecordsByUser(ctx context.Context, svcCtx *svc.ServiceContext, weeklyVoteID uint64, userID uint64) ([]types.VoteRecordInfo, error) {
	type voteRecordRow struct {
		Id           uint64    `db:"id"`
		CandidateId  uint64    `db:"candidate_id"`
		TopicId      uint64    `db:"topic_id"`
		TopicTitle   string    `db:"topic_title"`
		TopicSummary string    `db:"topic_summary"`
		CreatedAt    time.Time `db:"created_at"`
	}

	var rows []voteRecordRow
	query := `
		select r.id, r.candidate_id, c.topic_id, c.topic_title, c.topic_summary, r.created_at
		from weekly_topic_vote_records r
		join weekly_topic_vote_candidates c on c.id = r.candidate_id
		where r.weekly_vote_id = ? and r.user_id = ?
		order by r.created_at desc, r.id desc`
	if err := svcCtx.DB.QueryRowsCtx(ctx, &rows, query, weeklyVoteID, userID); err != nil {
		return nil, err
	}

	items := make([]types.VoteRecordInfo, 0, len(rows))
	for _, row := range rows {
		items = append(items, types.VoteRecordInfo{
			Id:           row.Id,
			CandidateId:  row.CandidateId,
			TopicId:      row.TopicId,
			TopicTitle:   row.TopicTitle,
			TopicSummary: row.TopicSummary,
			CreatedAt:    row.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return items, nil
}

func findTodayVoteRecord(records []types.VoteRecordInfo, now time.Time) (types.VoteRecordInfo, bool) {
	today := now.Format("2006-01-02")
	for _, record := range records {
		if strings.HasPrefix(record.CreatedAt, today) {
			return record, true
		}
	}

	return types.VoteRecordInfo{}, false
}

func weeklyVoteToInfo(bundle *weeklyVoteBundle) types.WeeklyVoteInfo {
	resultTopicID := uint64(0)
	if bundle.Vote.ResultTopicId.Valid {
		resultTopicID = uint64(bundle.Vote.ResultTopicId.Int64)
	}

	return types.WeeklyVoteInfo{
		Id:               bundle.Vote.Id,
		WeekStartDate:    bundle.Vote.WeekStartDate.Format("2006-01-02"),
		VoteStartAt:      bundle.Vote.VoteStartAt.Format("2006-01-02 15:04:05"),
		VoteEndAt:        bundle.Vote.VoteEndAt.Format("2006-01-02 15:04:05"),
		Status:           bundle.Vote.Status,
		ResultTopicId:    resultTopicID,
		HasVoted:         bundle.HasVoted,
		TodayHasVoted:    bundle.TodayHasVoted,
		UserCandidateId:  bundle.UserChoice,
		TodayCandidateId: bundle.TodayChoice,
		TodayVotedAt:     bundle.TodayVotedAt,
		Candidates:       bundle.Candidates,
		MyRecords:        bundle.MyRecords,
	}
}

func discussionToInfo(item *model.DiscussionInfos) types.DiscussionInfo {
	description := ""
	if item.Description.Valid {
		description = item.Description.String
	}

	goals := ""
	if item.Goals.Valid {
		goals = item.Goals.String
	}

	shareText := ""
	if item.ShareText.Valid {
		shareText = item.ShareText.String
	}

	adminRemark := ""
	if item.AdminRemark.Valid {
		adminRemark = item.AdminRemark.String
	}

	return types.DiscussionInfo{
		Id:              item.Id,
		WeekStartDate:   item.WeekStartDate.Format("2006-01-02"),
		TopicId:         item.TopicId,
		TopicTitle:      item.TopicTitle,
		DiscussionTitle: item.DiscussionTitle,
		Description:     description,
		Goals:           goals,
		MeetingTime:     item.MeetingTime.Format("2006-01-02 15:04:05"),
		MeetingLink:     item.MeetingLink,
		ShareText:       shareText,
		Status:          item.Status,
		AdminRemark:     adminRemark,
	}
}

func buildDerivedDiscussion(bundle *weeklyVoteBundle) types.DiscussionInfo {
	meetingTime := defaultDiscussionTime(bundle.Vote.WeekStartDate).Format("2006-01-02 15:04:05")
	topicTitle := "本周投票进行中"
	topicID := uint64(0)
	if bundle.Winner != nil {
		topicTitle = bundle.Winner.TopicTitle
		topicID = bundle.Winner.TopicId
	}

	discussion := types.DiscussionInfo{
		WeekStartDate:   bundle.Vote.WeekStartDate.Format("2006-01-02"),
		TopicId:         topicID,
		TopicTitle:      topicTitle,
		DiscussionTitle: defaultDiscussionTitle(topicTitle),
		Description:     defaultDiscussionDescription(topicTitle),
		Goals:           defaultDiscussionGoals(topicTitle),
		MeetingTime:     meetingTime,
		MeetingLink:     "",
		Status:          "published",
		AdminRemark:     "",
	}
	discussion.ShareText = buildDiscussionShareText(discussion)
	return discussion
}
