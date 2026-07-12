package logic

import (
	"context"
	"database/sql"
	"sort"
	"testing"
	"time"

	"api/internal/svc"
	"api/internal/types"
	"api/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type awarenessCheckInsertResult struct {
	id int64
}

func (r awarenessCheckInsertResult) LastInsertId() (int64, error) { return r.id, nil }
func (r awarenessCheckInsertResult) RowsAffected() (int64, error) { return 1, nil }

type awarenessCheckAwarenessModel struct {
	model.AwarenessModel
	points []model.Awareness
}

func (m *awarenessCheckAwarenessModel) FindEligible(context.Context) ([]model.Awareness, error) {
	return m.points, nil
}

func (m *awarenessCheckAwarenessModel) withSession(sqlx.Session) model.AwarenessModel { return m }

type awarenessCheckChaptersSourceModel struct {
	model.ChaptersModel
	chapters []model.Chapter
}

func (m *awarenessCheckChaptersSourceModel) FindAll(context.Context) ([]model.Chapter, error) {
	return m.chapters, nil
}

func (m *awarenessCheckChaptersSourceModel) withSession(sqlx.Session) model.ChaptersModel {
	return m
}

type awarenessCheckStore struct {
	model.AwarenessChecksModel
	nextID uint64
	items  map[uint64]*model.AwarenessCheck
}

func newAwarenessCheckStore() *awarenessCheckStore {
	return &awarenessCheckStore{
		nextID: 1,
		items:  map[uint64]*model.AwarenessCheck{},
	}
}

func (s *awarenessCheckStore) Insert(_ context.Context, data *model.AwarenessCheck) (sql.Result, error) {
	id := s.nextID
	s.nextID++
	item := *data
	item.CheckId = id
	if item.StartedAt.IsZero() {
		item.StartedAt = time.Date(2026, 7, int(id), 9, 0, 0, 0, time.Local)
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = item.StartedAt
	}
	if item.UpdatedAt.IsZero() {
		item.UpdatedAt = item.StartedAt
	}
	s.items[id] = &item
	return awarenessCheckInsertResult{id: int64(id)}, nil
}

func (s *awarenessCheckStore) FindOne(_ context.Context, id uint64) (*model.AwarenessCheck, error) {
	item, ok := s.items[id]
	if !ok {
		return nil, model.ErrNotFound
	}
	cloned := *item
	return &cloned, nil
}

func (s *awarenessCheckStore) FindCurrentByUserID(_ context.Context, userID uint64) (*model.AwarenessCheck, error) {
	var current *model.AwarenessCheck
	for _, item := range s.items {
		if item.UserId != userID {
			continue
		}
		if item.Status != checkStatusDraft && item.Status != checkStatusInProgress {
			continue
		}
		if current == nil || item.CheckId > current.CheckId {
			cloned := *item
			current = &cloned
		}
	}
	if current == nil {
		return nil, model.ErrNotFound
	}
	return current, nil
}

func (s *awarenessCheckStore) FindLatestDoneByUserID(_ context.Context, userID uint64) (*model.AwarenessCheck, error) {
	var latest *model.AwarenessCheck
	for _, item := range s.items {
		if item.UserId != userID || item.Status != checkStatusCompleted {
			continue
		}
		if latest == nil || item.CheckId > latest.CheckId {
			cloned := *item
			latest = &cloned
		}
	}
	if latest == nil {
		return nil, model.ErrNotFound
	}
	return latest, nil
}

func (s *awarenessCheckStore) FindLatestScoredByUserID(_ context.Context, userID uint64) (*model.AwarenessCheck, error) {
	var latest *model.AwarenessCheck
	for _, item := range s.items {
		if item.UserId != userID || item.Status == "abandoned" || !item.Score.Valid {
			continue
		}
		if latest == nil || item.CheckId > latest.CheckId {
			cloned := *item
			latest = &cloned
		}
	}
	if latest == nil {
		return nil, model.ErrNotFound
	}
	return latest, nil
}

func (s *awarenessCheckStore) FindByUserID(_ context.Context, userID uint64, limit int64) ([]model.AwarenessCheck, error) {
	var result []model.AwarenessCheck
	for _, item := range s.items {
		if item.UserId == userID {
			result = append(result, *item)
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].CheckId > result[j].CheckId
	})
	if limit > 0 && int64(len(result)) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (s *awarenessCheckStore) Update(_ context.Context, data *model.AwarenessCheck) error {
	item, ok := s.items[data.CheckId]
	if !ok {
		return model.ErrNotFound
	}
	cloned := *data
	cloned.UserId = item.UserId
	cloned.StartedAt = item.StartedAt
	cloned.CreatedAt = item.CreatedAt
	if cloned.UpdatedAt.IsZero() {
		cloned.UpdatedAt = time.Now()
	}
	s.items[data.CheckId] = &cloned
	return nil
}

func (s *awarenessCheckStore) withSession(sqlx.Session) model.AwarenessChecksModel { return s }

type awarenessCheckChapterStore struct {
	model.AwarenessCheckChaptersModel
	nextID uint64
	items  map[uint64]map[uint64]*model.AwarenessCheckChapter
	byID   map[uint64]*model.AwarenessCheckChapter
}

func newAwarenessCheckChapterStore() *awarenessCheckChapterStore {
	return &awarenessCheckChapterStore{
		nextID: 1,
		items:  map[uint64]map[uint64]*model.AwarenessCheckChapter{},
		byID:   map[uint64]*model.AwarenessCheckChapter{},
	}
}

func (s *awarenessCheckChapterStore) Insert(_ context.Context, data *model.AwarenessCheckChapter) (sql.Result, error) {
	id := s.nextID
	s.nextID++
	item := *data
	item.CheckChapterId = id
	if item.CreatedAt.IsZero() {
		item.CreatedAt = time.Now()
	}
	if item.UpdatedAt.IsZero() {
		item.UpdatedAt = item.CreatedAt
	}
	if s.items[item.CheckId] == nil {
		s.items[item.CheckId] = map[uint64]*model.AwarenessCheckChapter{}
	}
	s.items[item.CheckId][item.ChapterId] = &item
	s.byID[id] = &item
	return awarenessCheckInsertResult{id: int64(id)}, nil
}

func (s *awarenessCheckChapterStore) FindByCheckID(_ context.Context, checkID uint64) ([]model.AwarenessCheckChapter, error) {
	var result []model.AwarenessCheckChapter
	for _, item := range s.items[checkID] {
		result = append(result, *item)
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].ChapterId < result[j].ChapterId
	})
	return result, nil
}

func (s *awarenessCheckChapterStore) FindCompletedByUserID(_ context.Context, userID uint64, limit int64) ([]model.AwarenessCheckChapter, error) {
	var result []model.AwarenessCheckChapter
	for _, item := range s.byID {
		if item.UserId != userID || item.Status != checkStatusCompleted {
			continue
		}
		result = append(result, *item)
	}
	sort.SliceStable(result, func(i, j int) bool {
		left := result[i]
		right := result[j]
		if left.SubmittedAt.Valid && right.SubmittedAt.Valid && !left.SubmittedAt.Time.Equal(right.SubmittedAt.Time) {
			return left.SubmittedAt.Time.After(right.SubmittedAt.Time)
		}
		if left.CheckId == right.CheckId {
			return left.CheckChapterId > right.CheckChapterId
		}
		return left.CheckId > right.CheckId
	})
	if limit > 0 && int64(len(result)) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (s *awarenessCheckChapterStore) FindOneByCheckAndChapter(_ context.Context, checkID uint64, chapterID uint64) (*model.AwarenessCheckChapter, error) {
	if s.items[checkID] == nil || s.items[checkID][chapterID] == nil {
		return nil, model.ErrNotFound
	}
	cloned := *s.items[checkID][chapterID]
	return &cloned, nil
}

func (s *awarenessCheckChapterStore) Update(_ context.Context, data *model.AwarenessCheckChapter) error {
	item, ok := s.byID[data.CheckChapterId]
	if !ok {
		return model.ErrNotFound
	}
	cloned := *data
	cloned.CheckId = item.CheckId
	cloned.UserId = item.UserId
	cloned.ChapterId = item.ChapterId
	cloned.TotalPoints = item.TotalPoints
	cloned.CreatedAt = item.CreatedAt
	if cloned.UpdatedAt.IsZero() {
		cloned.UpdatedAt = time.Now()
	}
	s.items[cloned.CheckId][cloned.ChapterId] = &cloned
	s.byID[cloned.CheckChapterId] = &cloned
	return nil
}

func (s *awarenessCheckChapterStore) withSession(sqlx.Session) model.AwarenessCheckChaptersModel {
	return s
}

type awarenessCheckScoreStore struct {
	model.AwarenessCheckScoresModel
	nextID uint64
	items  map[uint64]map[uint64]*model.AwarenessCheckScore
}

func newAwarenessCheckScoreStore() *awarenessCheckScoreStore {
	return &awarenessCheckScoreStore{
		nextID: 1,
		items:  map[uint64]map[uint64]*model.AwarenessCheckScore{},
	}
}

func (s *awarenessCheckScoreStore) Upsert(_ context.Context, data *model.AwarenessCheckScore) (sql.Result, error) {
	if s.items[data.CheckId] == nil {
		s.items[data.CheckId] = map[uint64]*model.AwarenessCheckScore{}
	}
	item := *data
	if existing := s.items[data.CheckId][data.AwarenessId]; existing != nil {
		item.ScoreId = existing.ScoreId
		item.CreatedAt = existing.CreatedAt
	} else {
		item.ScoreId = s.nextID
		s.nextID++
		item.CreatedAt = time.Now()
	}
	item.UpdatedAt = time.Now()
	s.items[data.CheckId][data.AwarenessId] = &item
	return awarenessCheckInsertResult{id: int64(item.ScoreId)}, nil
}

func (s *awarenessCheckScoreStore) FindByCheckID(_ context.Context, checkID uint64) ([]model.AwarenessCheckScore, error) {
	var result []model.AwarenessCheckScore
	for _, item := range s.items[checkID] {
		result = append(result, *item)
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].ChapterId == result[j].ChapterId {
			return result[i].AwarenessId < result[j].AwarenessId
		}
		return result[i].ChapterId < result[j].ChapterId
	})
	return result, nil
}

func (s *awarenessCheckScoreStore) FindByCheckAndChapter(_ context.Context, checkID uint64, chapterID uint64) ([]model.AwarenessCheckScore, error) {
	var result []model.AwarenessCheckScore
	for _, item := range s.items[checkID] {
		if item.ChapterId == chapterID {
			result = append(result, *item)
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].AwarenessId < result[j].AwarenessId
	})
	return result, nil
}

func (s *awarenessCheckScoreStore) withSession(sqlx.Session) model.AwarenessCheckScoresModel {
	return s
}

func TestAwarenessCheckCurrentReusesExistingDraft(t *testing.T) {
	t.Parallel()

	svcCtx, checks := newAwarenessCheckTestService()
	ctx := WithCurrentUserID(context.Background(), 7)

	first, err := NewGetAwarenessCheckCurrentLogic(ctx, svcCtx).GetAwarenessCheckCurrent()
	if err != nil {
		t.Fatalf("unexpected current error: %v", err)
	}
	second, err := NewGetAwarenessCheckCurrentLogic(ctx, svcCtx).GetAwarenessCheckCurrent()
	if err != nil {
		t.Fatalf("unexpected second current error: %v", err)
	}

	if first.Data.Check.CheckId != second.Data.Check.CheckId {
		t.Fatalf("expected current check to be reused, got %d then %d", first.Data.Check.CheckId, second.Data.Check.CheckId)
	}
	if len(checks.items) != 1 {
		t.Fatalf("expected one current check, got %d", len(checks.items))
	}
	if len(first.Data.Chapters) != 1 || first.Data.Chapters[0].TotalPoints != 2 {
		t.Fatalf("expected one initialized chapter with two points, got %+v", first.Data.Chapters)
	}
}

func TestCreateAwarenessCheckAbandonsExistingCurrentRound(t *testing.T) {
	t.Parallel()

	svcCtx, checks := newAwarenessCheckTestService()
	ctx := WithCurrentUserID(context.Background(), 7)

	first, err := NewGetAwarenessCheckCurrentLogic(ctx, svcCtx).GetAwarenessCheckCurrent()
	if err != nil {
		t.Fatalf("unexpected current error: %v", err)
	}
	second, err := NewCreateAwarenessCheckLogic(ctx, svcCtx).CreateAwarenessCheck()
	if err != nil {
		t.Fatalf("unexpected create error: %v", err)
	}
	current, err := NewGetAwarenessCheckCurrentLogic(ctx, svcCtx).GetAwarenessCheckCurrent()
	if err != nil {
		t.Fatalf("unexpected current reload error: %v", err)
	}

	if first.Data.Check.CheckId == second.Data.Check.CheckId {
		t.Fatalf("expected a new round, got same check id %d", second.Data.Check.CheckId)
	}
	if current.Data.Check.CheckId != second.Data.Check.CheckId {
		t.Fatalf("expected current to be the new round, got current=%d new=%d", current.Data.Check.CheckId, second.Data.Check.CheckId)
	}
	oldCheck, err := checks.FindOne(ctx, first.Data.Check.CheckId)
	if err != nil {
		t.Fatalf("unexpected old check lookup error: %v", err)
	}
	if oldCheck.Status != checkStatusAbandoned {
		t.Fatalf("expected old current round to be abandoned, got %+v", oldCheck)
	}
}

func TestAwarenessCheckChapterCanBeRetestedAndKeepsHistoryChangesAndTrends(t *testing.T) {
	t.Parallel()

	svcCtx, _ := newAwarenessCheckTestService()
	ctx := WithCurrentUserID(context.Background(), 7)

	first, err := NewCreateAwarenessCheckLogic(ctx, svcCtx).CreateAwarenessCheck()
	if err != nil {
		t.Fatalf("unexpected create error: %v", err)
	}
	if first.Data.Check.CheckId != 1 || len(first.Data.MissingChapters) != 1 {
		t.Fatalf("expected first check with one missing chapter, got %+v", first.Data)
	}

	saveFirst, err := NewSaveAwarenessCheckChapterScoresLogic(ctx, svcCtx).SaveAwarenessCheckChapterScores(1, &types.AwarenessCheckScoreSaveReq{
		Scores: []types.AwarenessCheckScoreInput{
			{AwarenessId: 101, SelfScore: 40},
			{AwarenessId: 102, SelfScore: 60},
		},
	})
	if err != nil {
		t.Fatalf("unexpected first save error: %v", err)
	}
	if saveFirst.Data.Chapter.Status != checkChapterStatusInProgress || saveFirst.Data.Chapter.ScoredPoints != 2 {
		t.Fatalf("expected saved first chapter progress, got %+v", saveFirst.Data.Chapter)
	}
	if saveFirst.Data.Points[1].Direction != checkDirectionLower || saveFirst.Data.Points[1].Score != 40 {
		t.Fatalf("expected lower direction score to be reversed, got %+v", saveFirst.Data.Points[1])
	}

	firstDone, err := NewSubmitAwarenessCheckChapterLogic(ctx, svcCtx).SubmitAwarenessCheckChapter(1)
	if err != nil {
		t.Fatalf("unexpected first submit error: %v", err)
	}
	if firstDone.Data.Check.Status != checkStatusCompleted || firstDone.Data.Check.DoneChapters != 1 {
		t.Fatalf("expected first round completed, got %+v", firstDone.Data.Check)
	}
	if firstDone.Data.Chapter.Score != 40 || firstDone.Data.Chapter.RefScore != 65 {
		t.Fatalf("expected first chapter score 40/ref 65, got %+v", firstDone.Data.Chapter)
	}

	currentAfterFirst, err := NewGetAwarenessCheckCurrentLogic(ctx, svcCtx).GetAwarenessCheckCurrent()
	if err != nil {
		t.Fatalf("unexpected current after first submit error: %v", err)
	}
	if currentAfterFirst.Data.Check.DoneChapters != 1 || currentAfterFirst.Data.Chapters[0].Score != 40 {
		t.Fatalf("expected overview to show latest chapter result immediately, got %+v", currentAfterFirst.Data)
	}

	secondChapter, err := NewGetAwarenessCheckChapterLogic(ctx, svcCtx).GetAwarenessCheckChapter(1)
	if err != nil {
		t.Fatalf("unexpected second chapter error: %v", err)
	}
	if secondChapter.Data.Chapter.Status != checkChapterStatusNotStarted || secondChapter.Data.Chapter.PrevScore != 40 {
		t.Fatalf("expected chapter to be open for retest with previous score, got %+v", secondChapter.Data.Chapter)
	}
	if len(secondChapter.Data.Points) != 2 || !secondChapter.Data.Points[0].HasPrevScore {
		t.Fatalf("expected second chapter points with previous score, got %+v", secondChapter.Data.Points)
	}

	if _, err = NewSaveAwarenessCheckChapterScoresLogic(ctx, svcCtx).SaveAwarenessCheckChapterScores(1, &types.AwarenessCheckScoreSaveReq{
		Scores: []types.AwarenessCheckScoreInput{
			{AwarenessId: 101, SelfScore: 70},
			{AwarenessId: 102, SelfScore: 20},
		},
	}); err != nil {
		t.Fatalf("unexpected second save error: %v", err)
	}

	secondDone, err := NewSubmitAwarenessCheckChapterLogic(ctx, svcCtx).SubmitAwarenessCheckChapter(1)
	if err != nil {
		t.Fatalf("unexpected second submit error: %v", err)
	}
	if secondDone.Data.Check.Score != 75 || secondDone.Data.Compare.ScoreChange != 35 {
		t.Fatalf("expected second round score and overall change, got check=%+v compare=%+v", secondDone.Data.Check, secondDone.Data.Compare)
	}
	if secondDone.Data.Chapter.ScoreChange != 35 {
		t.Fatalf("expected chapter score change 35, got %+v", secondDone.Data.Chapter)
	}
	if len(secondDone.Data.Compare.ImprovedPoints) == 0 || secondDone.Data.Compare.ImprovedPoints[0].ScoreChange != 40 {
		t.Fatalf("expected improved point attribution, got %+v", secondDone.Data.Compare.ImprovedPoints)
	}

	history, err := NewListAwarenessCheckHistoryLogic(ctx, svcCtx).ListAwarenessCheckHistory()
	if err != nil {
		t.Fatalf("unexpected history error: %v", err)
	}
	if len(history.Data.List) != 2 || history.Data.List[0].CheckId != 2 || history.Data.List[1].CheckId != 1 {
		t.Fatalf("expected two history checks in reverse order, got %+v", history.Data.List)
	}

	trends, err := NewGetAwarenessCheckTrendsLogic(ctx, svcCtx).GetAwarenessCheckTrends()
	if err != nil {
		t.Fatalf("unexpected trends error: %v", err)
	}
	if len(trends.Data.Overall) != 2 || trends.Data.Overall[0].Score != 40 || trends.Data.Overall[1].Score != 75 {
		t.Fatalf("expected chronological overall trend, got %+v", trends.Data.Overall)
	}
	if len(trends.Data.Chapters) != 2 || trends.Data.Chapters[1].ScoreChange != 35 {
		t.Fatalf("expected chapter trend changes, got %+v", trends.Data.Chapters)
	}

	detail, err := NewGetAwarenessCheckDetailLogic(ctx, svcCtx).GetAwarenessCheckDetail(2)
	if err != nil {
		t.Fatalf("unexpected detail error: %v", err)
	}
	if len(detail.Data.Scores) != 2 || len(detail.Data.Compare.ImprovedChapters) != 1 {
		t.Fatalf("expected detail scores and chapter attribution, got %+v", detail.Data)
	}
}

func TestAwarenessCheckOverallTrendWaitsForEveryChapterLatestScore(t *testing.T) {
	t.Parallel()

	svcCtx, _ := newAwarenessCheckTestServiceWithData(
		[]model.Chapter{
			{ChapterId: 1, ChapterNo: 1, ChapterTitle: "第一章", ChapterFullTitle: "第一章 觉察起点", SortOrder: 1},
			{ChapterId: 2, ChapterNo: 2, ChapterTitle: "第二章", ChapterFullTitle: "第二章 稳定行动", SortOrder: 2},
		},
		[]model.Awareness{
			{AwarenessId: 101, ChapterId: 1, SectionId: 11, PointTitle: "看见状态", ReferenceMin: sql.NullFloat64{Float64: 50, Valid: true}, ReferenceMax: sql.NullFloat64{Float64: 50, Valid: true}, BetterDirection: checkDirectionHigher, SortOrderGlobal: 1, Status: 1},
			{AwarenessId: 201, ChapterId: 2, SectionId: 21, PointTitle: "稳定行动", ReferenceMin: sql.NullFloat64{Float64: 50, Valid: true}, ReferenceMax: sql.NullFloat64{Float64: 50, Valid: true}, BetterDirection: checkDirectionHigher, SortOrderGlobal: 2, Status: 1},
		},
	)
	ctx := WithCurrentUserID(context.Background(), 7)

	firstDone := saveAndSubmitAwarenessCheckChapter(t, ctx, svcCtx, 1, []types.AwarenessCheckScoreInput{
		{AwarenessId: 101, SelfScore: 60},
	})
	if firstDone.Data.Chapter.Score != 60 {
		t.Fatalf("expected first chapter score 60, got %+v", firstDone.Data.Chapter)
	}

	trendsAfterOne, err := NewGetAwarenessCheckTrendsLogic(ctx, svcCtx).GetAwarenessCheckTrends()
	if err != nil {
		t.Fatalf("unexpected trends after one chapter error: %v", err)
	}
	if len(trendsAfterOne.Data.Chapters) != 1 || len(trendsAfterOne.Data.Overall) != 0 {
		t.Fatalf("expected only chapter trend before every chapter has a score, got %+v", trendsAfterOne.Data)
	}

	secondDone := saveAndSubmitAwarenessCheckChapter(t, ctx, svcCtx, 2, []types.AwarenessCheckScoreInput{
		{AwarenessId: 201, SelfScore: 80},
	})
	if secondDone.Data.Chapter.Score != 80 {
		t.Fatalf("expected second chapter score 80, got %+v", secondDone.Data.Chapter)
	}

	trendsAfterAll, err := NewGetAwarenessCheckTrendsLogic(ctx, svcCtx).GetAwarenessCheckTrends()
	if err != nil {
		t.Fatalf("unexpected trends after all chapters error: %v", err)
	}
	if len(trendsAfterAll.Data.Chapters) != 2 || len(trendsAfterAll.Data.Overall) != 1 || trendsAfterAll.Data.Overall[0].Score != 70 {
		t.Fatalf("expected first overall trend after both chapters, got %+v", trendsAfterAll.Data)
	}

	thirdDone := saveAndSubmitAwarenessCheckChapter(t, ctx, svcCtx, 1, []types.AwarenessCheckScoreInput{
		{AwarenessId: 101, SelfScore: 90},
	})
	if thirdDone.Data.Chapter.ScoreChange != 30 {
		t.Fatalf("expected chapter retest change 30, got %+v", thirdDone.Data.Chapter)
	}

	trendsAfterRetest, err := NewGetAwarenessCheckTrendsLogic(ctx, svcCtx).GetAwarenessCheckTrends()
	if err != nil {
		t.Fatalf("unexpected trends after retest error: %v", err)
	}
	if len(trendsAfterRetest.Data.Chapters) != 3 || len(trendsAfterRetest.Data.Overall) != 2 {
		t.Fatalf("expected chapter retest and second overall trend point, got %+v", trendsAfterRetest.Data)
	}
	if trendsAfterRetest.Data.Overall[1].Score != 85 {
		t.Fatalf("expected latest overall score to use retested chapter, got %+v", trendsAfterRetest.Data.Overall)
	}
}

func saveAndSubmitAwarenessCheckChapter(t *testing.T, ctx context.Context, svcCtx *svc.ServiceContext, chapterID uint64, scores []types.AwarenessCheckScoreInput) *types.AwarenessCheckChapterResp {
	t.Helper()
	if _, err := NewSaveAwarenessCheckChapterScoresLogic(ctx, svcCtx).SaveAwarenessCheckChapterScores(chapterID, &types.AwarenessCheckScoreSaveReq{
		Scores: scores,
	}); err != nil {
		t.Fatalf("unexpected save chapter %d error: %v", chapterID, err)
	}
	done, err := NewSubmitAwarenessCheckChapterLogic(ctx, svcCtx).SubmitAwarenessCheckChapter(chapterID)
	if err != nil {
		t.Fatalf("unexpected submit chapter %d error: %v", chapterID, err)
	}
	return done
}

func newAwarenessCheckTestService() (*svc.ServiceContext, *awarenessCheckStore) {
	return newAwarenessCheckTestServiceWithData(
		[]model.Chapter{
			{
				ChapterId:        1,
				ChapterNo:        1,
				ChapterTitle:     "第一章",
				ChapterFullTitle: "第一章 觉察起点",
				SortOrder:        1,
			},
		},
		[]model.Awareness{
			{
				AwarenessId:     101,
				ChapterId:       1,
				SectionId:       11,
				PointTitle:      "看见状态",
				Summary:         sql.NullString{String: "先看见自己的状态", Valid: true},
				ReferenceMin:    sql.NullFloat64{Float64: 50, Valid: true},
				ReferenceMax:    sql.NullFloat64{Float64: 50, Valid: true},
				BetterDirection: checkDirectionHigher,
				SortOrderGlobal: 1,
				Status:          1,
			},
			{
				AwarenessId:     102,
				ChapterId:       1,
				SectionId:       12,
				PointTitle:      "降低反应",
				Summary:         sql.NullString{String: "反应越低越稳定", Valid: true},
				ReferenceMin:    sql.NullFloat64{Float64: 20, Valid: true},
				ReferenceMax:    sql.NullFloat64{Float64: 20, Valid: true},
				BetterDirection: checkDirectionLower,
				SortOrderGlobal: 2,
				Status:          1,
			},
		},
	)
}

func newAwarenessCheckTestServiceWithData(chapters []model.Chapter, points []model.Awareness) (*svc.ServiceContext, *awarenessCheckStore) {
	checks := newAwarenessCheckStore()
	return &svc.ServiceContext{
		AwarenessModel: &awarenessCheckAwarenessModel{
			points: points,
		},
		ChaptersModel: &awarenessCheckChaptersSourceModel{
			chapters: chapters,
		},
		AwarenessChecksModel:        checks,
		AwarenessCheckChaptersModel: newAwarenessCheckChapterStore(),
		AwarenessCheckScoresModel:   newAwarenessCheckScoreStore(),
	}, checks
}
