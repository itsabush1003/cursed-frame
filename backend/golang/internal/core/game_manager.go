package core

import (
	"context"
	"errors"
	"math"
	"slices"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/util"
)

type lobby struct {
	users        []uuid.UUID
	ctx          context.Context
	doneNotifier context.CancelFunc
}

func (l *lobby) Join(user uuid.UUID) {
	l.users = append(l.users, user)
}

func (l *lobby) Disconnect(user uuid.UUID) {
	idx := slices.Index(l.users, user)
	if idx > 0 {
		l.users = append(l.users[:idx], l.users[idx+1:]...)
	} else if idx == 0 {
		l.users = l.users[1:]
	}
}

type TeamID uint32

type Choice struct {
	Target     uuid.UUID
	ChoiceID   uint
	ChoiceText string
}

type AnswerWithMap struct {
	TeamAnswer Choice
	AnswerMap  map[uint]int
}

type Quiz struct {
	ImageID      string
	TeamID       TeamID
	QuestionID   uint
	QuestionText string
	Choices      []Choice
	RemainedTime int
}

const (
	MaxChoiceNum          int = 4
	MaxHintLength         int = 30
	InitialRemaindTime    int = 15
	IncreaseTimeHintTaken int = 10
)

type questRoom struct {
	teams            map[TeamID][]uuid.UUID
	conn             map[uuid.UUID]chan<- Quiz
	hintCh           chan string
	answerListener   map[TeamID]chan Choice
	answerSender     map[uuid.UUID]chan AnswerWithMap
	nextQuizNotifier chan struct{}
	mu               sync.RWMutex
	ctx              context.Context
	doneNotifier     context.CancelFunc
	currentTarget    uuid.UUID
	currentAnswer    Choice
	quizCount        int
	teamStats        map[TeamID]int
	personalStats    map[uuid.UUID]int
}

func (qr *questRoom) SetCurrent(target uuid.UUID, answer Choice) {
	qr.mu.Lock()
	defer qr.mu.Unlock()
	qr.currentTarget = target
	qr.currentAnswer = answer
}

func (qr *questRoom) PublishQuiz(quiz Quiz) {
	qr.mu.RLock()
	defer qr.mu.RUnlock()
	for _, con := range qr.conn {
		if con == nil {
			continue
		}
		go func(ch chan<- Quiz) {
			select {
			case ch <- quiz:
				return
			case <-time.After(time.Second):
				return
			case <-qr.ctx.Done():
				return
			}
		}(con)
	}
}

func (qr *questRoom) CollectAnswer() (map[TeamID]Choice, map[TeamID]map[uint]int) {
	var wg sync.WaitGroup
	qr.mu.RLock()
	defer qr.mu.RUnlock()
	reporters := make(map[TeamID]chan Choice, len(qr.teams))
	for tid, _ := range qr.teams {
		reporters[tid] = make(chan Choice, len(qr.teams[tid]))
		wg.Go(func() {
			timer := time.NewTimer(5 * time.Second)
			defer timer.Stop()
			defer close(reporters[tid])
			for {
				select {
				case answer := <-qr.answerListener[tid]:
					reporters[tid] <- answer
					// チャネルのバッファにチーム人数分の回答が溜まっている
					// ＝ チームの回答が出揃った
					// のでこれ以上の回収はせず終了する
					if len(reporters[tid]) == len(qr.teams[tid]) {
						return
					}
				case <-qr.ctx.Done():
					return
				case <-timer.C: // Timeout
					return
				}
			}
		})
	}

	wg.Wait()
	mu := sync.Mutex{}
	teamAnswers := make(map[TeamID]Choice, len(qr.teams))
	teamAnswersMap := make(map[TeamID]map[uint]int, len(qr.teams))
	for tid, answers := range reporters {
		wg.Go(func() {
			res := make([]Choice, 0, len(answers))
			choiceCounter := make(map[uint]int, MaxChoiceNum)
			for answer := range answers {
				res = append(res, answer)
				choiceCounter[answer.ChoiceID]++
			}
			var maxChoiceCnt int = 0
			var maxChoiceIDs []uint = make([]uint, 0, MaxChoiceNum)
			for cid, cnt := range choiceCounter {
				if cnt == 0 {
					continue
				}
				if cnt > maxChoiceCnt {
					maxChoiceCnt = cnt
					maxChoiceIDs = append(maxChoiceIDs[:0], cid)
				} else if cnt == maxChoiceCnt {
					maxChoiceIDs = append(maxChoiceIDs, cid)
				}
			}
			if len(maxChoiceIDs) > 1 {
				maxChoiceIDs = util.ShuffleSlice(maxChoiceIDs)
			}
			for _, ans := range res {
				if ans.ChoiceID == maxChoiceIDs[0] {
					mu.Lock()
					defer mu.Unlock()
					teamAnswers[tid] = ans
					teamAnswersMap[tid] = choiceCounter
					break
				}
			}
		})
	}

	wg.Wait()
	return teamAnswers, teamAnswersMap
}

func (qr *questRoom) UpdateTeamStats(teamAnswers map[TeamID]Choice) {
	qr.mu.Lock()
	defer qr.mu.Unlock()
	qr.quizCount++
	for tid, choice := range teamAnswers {
		if choice.ChoiceID == qr.currentAnswer.ChoiceID {
			qr.teamStats[tid]++
		}
	}
}

func (qr *questRoom) Connect(uid uuid.UUID) (context.Context, <-chan Quiz) {
	qr.mu.Lock()
	defer qr.mu.Unlock()
	if qr.conn[uid] != nil {
		close(qr.conn[uid])
	}
	ch := make(chan Quiz, 1)
	qr.conn[uid] = ch
	return qr.ctx, ch
}

func (qr *questRoom) Answer(tid TeamID, uid uuid.UUID, answer Choice) AnswerWithMap {
	qr.mu.RLock()
	defer qr.mu.RUnlock()
	qr.answerListener[tid] <- answer
	teamAnswer := <-qr.answerSender[uid]
	return teamAnswer
}

func (qr *questRoom) UpdatePersonalStats(uid uuid.UUID, answer Choice) {
	qr.mu.Lock()
	defer qr.mu.Unlock()
	if answer.ChoiceID == qr.currentAnswer.ChoiceID {
		qr.personalStats[uid]++
	}
}

type State int

const (
	UNDEFINED State = iota
	INITIALIZED
	ACCEPTING
	CLOSED
	INGAME
	RESULT
)

type Result struct {
	Answer    Choice
	IsCorrect bool
}

type Stats struct {
	CorrectRate float32
	Order       int
}

type GameManager struct {
	state      State
	maxUserNum int
	teamNum    int
	ctx        context.Context
	mu         sync.RWMutex
	lobby      *lobby
	room       *questRoom
}

func (gm *GameManager) OpenLobby() (context.Context, error) {
	if gm.state != INITIALIZED && gm.state != ACCEPTING {
		return nil, errors.New("Lobby has already been opend before")
	}
	gm.mu.Lock()
	defer gm.mu.Unlock()
	gm.state = ACCEPTING
	return gm.lobby.ctx, nil
}

func (gm *GameManager) CloseLobby() error {
	if gm.state != ACCEPTING && gm.state != CLOSED {
		return errors.New("Lobby has not opend yet, or already closed")
	}
	gm.mu.Lock()
	defer gm.mu.Unlock()
	gm.state = CLOSED
	return nil
}

func (gm *GameManager) NotifyLobbyClosed() error {
	if gm.state != CLOSED {
		return errors.New("Lobby has not opend yet, or already closed")
	}
	gm.lobby.doneNotifier()
	return nil
}

func (gm *GameManager) GetLobbyUsers() []uuid.UUID {
	if gm.state == ACCEPTING || gm.state == CLOSED {
		gm.mu.RLock()
		defer gm.mu.RUnlock()
		lobbyUsers := make([]uuid.UUID, len(gm.lobby.users))
		copy(lobbyUsers, gm.lobby.users)
		return lobbyUsers
	} else {
		return nil
	}
}

func (gm *GameManager) SplitTeams(users []uuid.UUID, teamNum int) map[uuid.UUID]uint32 {
	if gm.state != CLOSED {
		return nil
	}

	gm.mu.Lock()
	defer gm.mu.Unlock()
	userNum := len(users)
	userTeam := make(map[uuid.UUID]uint32, userNum)
	maxTeamMember := int(math.Ceil(float64(userNum) / float64(teamNum)))
	for i := range teamNum {
		gm.room.teams[TeamID(i)] = make([]uuid.UUID, 0, maxTeamMember)
	}
	shuffled := util.ShuffleSlice(users)
	for i, u := range shuffled {
		userTeam[u] = uint32((i % teamNum) + 1)
	}
	for u, i := range userTeam {
		gm.room.teams[TeamID(i)] = append(gm.room.teams[TeamID(i)], u)
	}
	return userTeam
}

func (gm *GameManager) GetTeams() map[TeamID][]uuid.UUID {
	if gm.state < CLOSED {
		return nil
	}
	teams := make(map[TeamID][]uuid.UUID)
	for tid, uidList := range gm.room.teams {
		teams[tid] = make([]uuid.UUID, len(uidList))
		copy(teams[tid], uidList)
	}
	return teams
}

func (gm *GameManager) QuestStart() (<-chan struct{}, error) {
	if gm.state != CLOSED && gm.state != INGAME {
		return nil, errors.New("Not quest ready or quest has already done")
	}
	gm.mu.Lock()
	defer gm.mu.Unlock()
	gm.state = INGAME
	for tid, uids := range gm.room.teams {
		gm.room.answerListener[tid] = make(chan Choice)
		for _, uid := range uids {
			gm.room.answerSender[uid] = make(chan AnswerWithMap)
		}
	}
	return gm.room.nextQuizNotifier, nil
}

func (gm *GameManager) Broadcast(target uuid.UUID, quiz Quiz, correct Choice) error {
	if gm.state != INGAME {
		return errors.New("Server is not in game mode")
	}
	gm.mu.Lock()
	defer gm.mu.Unlock()
	gm.room.SetCurrent(target, correct)
	gm.room.PublishQuiz(quiz)
	return nil
}

func (gm *GameManager) CheckHint() <-chan string {
	return gm.room.hintCh
}

func (gm *GameManager) CollectAnswer() (map[TeamID]Result, map[TeamID]map[uint]int, error) {
	if gm.state != INGAME {
		return nil, nil, errors.New("Server is not in game mode")
	}
	teamAnswers, teamAnswersMap := gm.room.CollectAnswer()
	gm.room.UpdateTeamStats(teamAnswers)
	results := make(map[TeamID]Result, len(teamAnswers))
	for tid, choice := range teamAnswers {
		results[tid] = Result{
			Answer:    choice,
			IsCorrect: choice == gm.room.currentAnswer,
		}
	}
	return results, teamAnswersMap, nil
}

func (gm *GameManager) DistributeAnswer(results map[TeamID]Result, answerMaps map[TeamID]map[uint]int) error {
	if gm.state != INGAME {
		return errors.New("Server is not in game mode")
	}

	for tid, result := range results {
		users := gm.room.teams[tid]
		for _, user := range users {
			go func(ch chan<- AnswerWithMap, c Choice, am map[uint]int) {
				select {
				case ch <- AnswerWithMap{
					TeamAnswer: c,
					AnswerMap:  am,
				}:
					return
				case <-time.After(time.Second):
					return
				}
			}(gm.room.answerSender[user], result.Answer, answerMaps[tid])
		}
	}
	return nil
}

func (gm *GameManager) NextQuiz() error {
	if gm.state != INGAME {
		return errors.New("Server is not in game mode")
	}

	select {
	case gm.room.nextQuizNotifier <- struct{}{}:
		return nil
	case <-time.After(time.Second):
		return errors.New("Timed out")
	}
}

func (gm *GameManager) EndQuest() error {
	if gm.state != INGAME {
		return errors.New("Server is not in game mode")
	}

	gm.mu.Lock()
	defer gm.mu.Unlock()
	gm.state = RESULT
	return nil
}

func (gm *GameManager) GetAllStats() (float32, map[uuid.UUID]Stats, map[TeamID]Stats, error) {
	if gm.state != RESULT {
		return 0.0, nil, nil, errors.New("Game has not been ended")
	}

	gm.mu.RLock()
	defer gm.mu.RUnlock()
	usersStats := make(map[uuid.UUID]Stats, len(gm.room.personalStats))
	usersCounts := make([]int, 0, len(gm.room.personalStats))
	for _, cnt := range gm.room.personalStats {
		usersCounts = append(usersCounts, cnt)
	}
	// descで並べたいので
	slices.SortFunc(usersCounts, func(a, b int) int { return b - a })
	for uid, cnt := range gm.room.personalStats {
		usersStats[uid] = Stats{
			CorrectRate: float32(cnt) / float32(gm.room.quizCount),
			Order:       slices.Index(usersCounts, cnt) + 1,
		}
	}
	teamsStats := make(map[TeamID]Stats, len(gm.room.teamStats))
	teamsCounts := make([]int, 0, len(gm.room.teamStats))
	var sum int = 0
	for _, cnt := range gm.room.teamStats {
		teamsCounts = append(teamsCounts, cnt)
		sum += cnt
	}
	// descで並べたいので
	slices.SortFunc(teamsCounts, func(a, b int) int { return b - a })
	for uid, cnt := range gm.room.teamStats {
		teamsStats[uid] = Stats{
			CorrectRate: float32(cnt) / float32(gm.room.quizCount),
			Order:       slices.Index(teamsCounts, cnt) + 1,
		}
	}
	return float32(sum) / float32(gm.room.quizCount*len(gm.room.teamStats)), usersStats, teamsStats, nil
}

func (gm *GameManager) JoinLobby(uid uuid.UUID) (context.Context, error) {
	if gm.state != ACCEPTING {
		return nil, errors.New("Server is not accepting now")
	}
	gm.mu.Lock()
	defer gm.mu.Unlock()
	gm.lobby.Join(uid)
	return gm.lobby.ctx, nil
}

func (gm *GameManager) DisconnectLobby(uid uuid.UUID) error {
	if gm.state >= INGAME {
		return errors.New("Lobby has already empty")
	}
	gm.mu.Lock()
	defer gm.mu.Unlock()
	gm.lobby.Disconnect(uid)
	return nil
}

func (gm *GameManager) EnterQuestRoom(uid uuid.UUID) (context.Context, <-chan Quiz, error) {
	if gm.state != INGAME {
		return nil, nil, errors.New("Not open the quest room")
	}
	gm.mu.Lock()
	defer gm.mu.Unlock()
	ctx, ch := gm.room.Connect(uid)
	return ctx, ch, nil
}

func (gm *GameManager) TakeHint(uid uuid.UUID, hint string) error {
	if gm.state != INGAME {
		return errors.New("Game is not start or has ended")
	}
	if gm.room.currentTarget != uid {
		return errors.New("You cannot take a hint")
	}
	if utf8.RuneCountInString(hint) > MaxHintLength {
		return errors.New("Your hint is too long")
	}
	select {
	case gm.room.hintCh <- hint:
		return nil
	case <-time.After(2 * time.Second):
		return errors.New("Send hint has timed out")
	}
}

func (gm *GameManager) Answer(uid uuid.UUID, tid TeamID, answer Choice) (AnswerWithMap, Choice, error) {
	if gm.state != INGAME {
		return AnswerWithMap{}, Choice{}, errors.New("Game is not start or has ended")
	}
	teamAnswer := gm.room.Answer(tid, uid, answer)
	gm.room.UpdatePersonalStats(uid, answer)
	return teamAnswer, gm.room.currentAnswer, nil
}

func (gm *GameManager) GetResultStats(uid uuid.UUID, tid TeamID) (total float32, ps Stats, ts Stats, err error) {
	if gm.state != RESULT {
		return 0.0, Stats{}, Stats{}, errors.New("Game has not been ended")
	}
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	psSorts := make([]int, 0, len(gm.room.personalStats))
	for _, cnt := range gm.room.personalStats {
		psSorts = append(psSorts, cnt)
	}
	// descで並べたいので
	slices.SortFunc(psSorts, func(a, b int) int { return b - a })
	ps = Stats{
		CorrectRate: float32(gm.room.personalStats[uid]) / float32(gm.room.quizCount),
		Order:       slices.Index(psSorts, gm.room.personalStats[uid]) + 1,
	}
	tsSorts := make([]int, 0, len(gm.room.teamStats))
	var sum int = 0
	for _, cnt := range gm.room.teamStats {
		tsSorts = append(tsSorts, cnt)
		sum += cnt
	}
	// descで並べたいので
	slices.SortFunc(tsSorts, func(a, b int) int { return b - a })
	ts = Stats{
		CorrectRate: float32(gm.room.teamStats[tid]) / float32(gm.room.quizCount),
		Order:       slices.Index(tsSorts, gm.room.teamStats[tid]) + 1,
	}
	return float32(sum) / float32(gm.room.quizCount*len(gm.room.teamStats)), ps, ts, nil
}

func NewGameManager(maxUserNum int, teamNum int) *GameManager {
	return sync.OnceValue(func() *GameManager {
		rootCtx := context.Background()
		lobbyCtx, lobbyDone := context.WithCancel(rootCtx)
		roomCtx, roomDone := context.WithCancel(rootCtx)
		return &GameManager{
			state:      INITIALIZED,
			maxUserNum: maxUserNum,
			teamNum:    teamNum,
			ctx:        rootCtx,
			mu:         sync.RWMutex{},
			lobby: &lobby{
				users:        make([]uuid.UUID, 0, maxUserNum),
				ctx:          lobbyCtx,
				doneNotifier: lobbyDone,
			},
			room: &questRoom{
				teams:            make(map[TeamID][]uuid.UUID, teamNum),
				conn:             make(map[uuid.UUID]chan<- Quiz, maxUserNum),
				hintCh:           make(chan string),
				answerListener:   make(map[TeamID]chan Choice, teamNum),
				answerSender:     make(map[uuid.UUID]chan AnswerWithMap, maxUserNum),
				nextQuizNotifier: make(chan struct{}),
				mu:               sync.RWMutex{},
				ctx:              roomCtx,
				doneNotifier:     roomDone,
				quizCount:        0,
				teamStats:        make(map[TeamID]int, teamNum),
				personalStats:    make(map[uuid.UUID]int, maxUserNum),
			},
		}
	})()
}
