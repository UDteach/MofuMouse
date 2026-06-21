package core

import (
	"strings"
	"sync/atomic"
	"time"
)

type KeepAwakeMode string

const (
	KeepAwakeOff KeepAwakeMode = "off"
	KeepAwakeOS  KeepAwakeMode = "os"
)

type JiggleMode string

const (
	JiggleOff JiggleMode = "off"
	Jiggle1px JiggleMode = "1px"
)

type CoatColor string

const (
	CoatAgouti CoatColor = "agouti"
	CoatGray   CoatColor = "gray"
	CoatDark   CoatColor = "dark"
	CoatCream  CoatColor = "cream"
	CoatWhite  CoatColor = "white"
	CoatPied   CoatColor = "pied"
)

type Settings struct {
	PetName            string
	AssistantEnabled   bool
	LaunchAtLogin      bool
	QuickKeysEnabled   bool
	QuickKeys          []QuickKeyID
	QuickScrollEnabled bool
	QuickScrollLines   int
	TargetAssist       bool
	TargetMove         bool
	TargetReturn       bool
	TargetMoveDelayMs  int
	TargetRules        []AssistRule
	KeepAwakeMode      KeepAwakeMode
	JiggleMode         JiggleMode
	JiggleIdleSec      int
	SpriteSizePx       int
	ShowPetName        bool
	CoatColor          CoatColor
	CompanionOffsetX   int
	CompanionOffsetY   int
	EdgeWarpEnabled    bool
	IdleActions        bool
	IdleActionSec      int
	UpdateCheck        bool
	UpdateRepo         string
}

func DefaultSettings() Settings {
	return Settings{
		PetName:            "Mofu",
		AssistantEnabled:   true,
		LaunchAtLogin:      false,
		QuickKeysEnabled:   true,
		QuickKeys:          DefaultQuickKeys(),
		QuickScrollEnabled: true,
		QuickScrollLines:   3,
		TargetAssist:       false,
		TargetMove:         false,
		TargetReturn:       false,
		TargetMoveDelayMs:  900,
		KeepAwakeMode:      KeepAwakeOff,
		JiggleMode:         JiggleOff,
		JiggleIdleSec:      60,
		SpriteSizePx:       32,
		ShowPetName:        false,
		CoatColor:          CoatAgouti,
		CompanionOffsetX:   0,
		CompanionOffsetY:   0,
		EdgeWarpEnabled:    false,
		IdleActions:        true,
		IdleActionSec:      45,
		UpdateCheck:        false,
		UpdateRepo:         "UDteach/MofuMouse",
	}
}

type State struct {
	assistantEnabled   atomic.Bool
	launchAtLogin      atomic.Bool
	quickKeysEnabled   atomic.Bool
	quickScrollEnabled atomic.Bool
	targetAssist       atomic.Bool
	targetMove         atomic.Bool
	targetReturn       atomic.Bool
	keepAwakeOS        atomic.Bool
	jiggle1px          atomic.Bool
	edgeWarpEnabled    atomic.Bool
	idleActions        atomic.Bool
	updateCheck        atomic.Bool
	showPetName        atomic.Bool

	petName     atomic.Value
	coatColor   atomic.Value
	quickKeys   atomic.Value
	targetRules atomic.Value
	spriteSize  atomic.Int32
	idleAction  atomic.Value
	actionUntil atomic.Int64

	JiggleIdleSec     int
	IdleActionSec     int
	QuickScrollLines  int
	TargetMoveDelayMs int
	CompanionOffsetX  int
	CompanionOffsetY  int
	UpdateRepo        string
}

func NewState(settings Settings) *State {
	defaults := DefaultSettings()
	if settings.JiggleIdleSec <= 0 {
		settings.JiggleIdleSec = defaults.JiggleIdleSec
	}
	if settings.IdleActionSec <= 0 {
		settings.IdleActionSec = defaults.IdleActionSec
	}
	settings.TargetMoveDelayMs = sanitizeTargetMoveDelayMs(settings.TargetMoveDelayMs)
	settings.QuickScrollLines = SanitizeQuickScrollLines(settings.QuickScrollLines)
	if strings.TrimSpace(settings.UpdateRepo) == "" {
		settings.UpdateRepo = defaults.UpdateRepo
	}

	s := &State{
		JiggleIdleSec:     settings.JiggleIdleSec,
		IdleActionSec:     settings.IdleActionSec,
		QuickScrollLines:  settings.QuickScrollLines,
		TargetMoveDelayMs: settings.TargetMoveDelayMs,
		UpdateRepo:        strings.TrimSpace(settings.UpdateRepo),
	}
	s.CompanionOffsetX, s.CompanionOffsetY = sanitizeCompanionOffsets(settings.CompanionOffsetX, settings.CompanionOffsetY)
	s.petName.Store(sanitizePetName(settings.PetName))
	s.coatColor.Store(sanitizeCoatColor(settings.CoatColor))
	s.assistantEnabled.Store(settings.AssistantEnabled)
	s.launchAtLogin.Store(settings.LaunchAtLogin)
	s.quickKeysEnabled.Store(settings.QuickKeysEnabled)
	s.quickKeys.Store(SanitizeQuickKeys(settings.QuickKeys))
	s.quickScrollEnabled.Store(settings.QuickScrollEnabled)
	s.targetAssist.Store(false)
	s.targetMove.Store(false)
	s.targetReturn.Store(false)
	s.targetRules.Store([]AssistRule{})
	s.keepAwakeOS.Store(false)
	s.jiggle1px.Store(settings.JiggleMode == Jiggle1px)
	s.edgeWarpEnabled.Store(settings.EdgeWarpEnabled)
	s.idleActions.Store(settings.IdleActions)
	s.updateCheck.Store(settings.UpdateCheck)
	s.showPetName.Store(settings.ShowPetName)
	s.SetSpriteSizePx(settings.SpriteSizePx)
	s.idleAction.Store("")
	return s
}

func (s *State) ApplySettings(settings Settings) {
	s.SetName(settings.PetName)
	s.SetAssistantEnabled(settings.AssistantEnabled)
	s.SetLaunchAtLogin(settings.LaunchAtLogin)
	s.SetQuickKeysEnabled(settings.QuickKeysEnabled)
	s.SetQuickKeys(settings.QuickKeys)
	s.SetQuickScrollEnabled(settings.QuickScrollEnabled)
	s.SetQuickScrollLines(settings.QuickScrollLines)
	s.SetTargetAssistEnabled(false)
	s.SetTargetMoveEnabled(false)
	s.SetTargetReturnEnabled(false)
	s.SetTargetMoveDelayMs(settings.TargetMoveDelayMs)
	s.SetTargetRules(nil)
	s.SetKeepAwakeOS(false)
	s.SetJiggle1px(settings.JiggleMode == Jiggle1px)
	s.SetJiggleIdleSec(settings.JiggleIdleSec)
	s.SetEdgeWarpEnabled(settings.EdgeWarpEnabled)
	s.SetIdleActionsEnabled(settings.IdleActions)
	s.SetIdleActionSec(settings.IdleActionSec)
	s.SetSpriteSizePx(settings.SpriteSizePx)
	s.SetShowPetName(settings.ShowPetName)
	s.SetCoatColor(settings.CoatColor)
	s.SetCompanionOffsets(settings.CompanionOffsetX, settings.CompanionOffsetY)
	s.SetUpdateCheckEnabled(settings.UpdateCheck)
	s.SetUpdateRepo(settings.UpdateRepo)
}

func sanitizeCompanionOffsets(x, y int) (int, int) {
	if (x == 4 && y == -22) || (x == 28 && y == 22) {
		return 0, 0
	}
	return clampInt(x, -48, 48), clampInt(y, -48, 48)
}

func clampInt(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func (s *State) Name() string {
	name, _ := s.petName.Load().(string)
	if name == "" {
		return DefaultSettings().PetName
	}
	return name
}

func (s *State) SetName(name string) {
	s.petName.Store(sanitizePetName(name))
}

func (s *State) CoatColor() CoatColor {
	coat, _ := s.coatColor.Load().(CoatColor)
	return sanitizeCoatColor(coat)
}

func (s *State) SetCoatColor(coat CoatColor) {
	s.coatColor.Store(sanitizeCoatColor(coat))
}

func (s *State) AssistantEnabled() bool {
	return s.assistantEnabled.Load()
}

func (s *State) SetAssistantEnabled(enabled bool) {
	s.assistantEnabled.Store(enabled)
}

func (s *State) ToggleAssistant() bool {
	return toggleBool(&s.assistantEnabled)
}

func (s *State) LaunchAtLogin() bool {
	return s.launchAtLogin.Load()
}

func (s *State) SetLaunchAtLogin(enabled bool) {
	s.launchAtLogin.Store(enabled)
}

func (s *State) ToggleLaunchAtLogin() bool {
	return toggleBool(&s.launchAtLogin)
}

func (s *State) QuickKeysEnabled() bool {
	return s.quickKeysEnabled.Load()
}

func (s *State) SetQuickKeysEnabled(enabled bool) {
	s.quickKeysEnabled.Store(enabled)
}

func (s *State) ToggleQuickKeysEnabled() bool {
	return toggleBool(&s.quickKeysEnabled)
}

func (s *State) QuickKeys() []QuickKeyID {
	keys, _ := s.quickKeys.Load().([]QuickKeyID)
	return SanitizeQuickKeys(keys)
}

func (s *State) SetQuickKeys(keys []QuickKeyID) {
	s.quickKeys.Store(SanitizeQuickKeys(keys))
}

func (s *State) QuickKeyAllowed(key QuickKeyID) bool {
	if !s.QuickKeysEnabled() {
		return false
	}
	for _, configured := range s.QuickKeys() {
		if configured == key {
			return true
		}
	}
	return false
}

func (s *State) QuickScrollEnabled() bool {
	return s.quickScrollEnabled.Load()
}

func (s *State) SetQuickScrollEnabled(enabled bool) {
	s.quickScrollEnabled.Store(enabled)
}

func (s *State) ToggleQuickScrollEnabled() bool {
	return toggleBool(&s.quickScrollEnabled)
}

func (s *State) SetQuickScrollLines(lines int) {
	s.QuickScrollLines = SanitizeQuickScrollLines(lines)
}

func (s *State) QuickScrollAllowed(direction QuickScrollDirection) bool {
	return s.QuickScrollEnabled() && IsAllowedQuickScrollDirection(direction)
}

func (s *State) TargetAssistEnabled() bool {
	return false
}

func (s *State) SetTargetAssistEnabled(enabled bool) {
	s.targetAssist.Store(false)
}

func (s *State) ToggleTargetAssist() bool {
	s.SetTargetAssistEnabled(false)
	return false
}

func (s *State) TargetMoveEnabled() bool {
	return false
}

func (s *State) SetTargetMoveEnabled(enabled bool) {
	s.targetMove.Store(false)
}

func (s *State) ToggleTargetMove() bool {
	s.SetTargetMoveEnabled(false)
	return false
}

func (s *State) TargetReturnEnabled() bool {
	return false
}

func (s *State) SetTargetReturnEnabled(enabled bool) {
	s.targetReturn.Store(false)
}

func (s *State) ToggleTargetReturn() bool {
	s.SetTargetReturnEnabled(false)
	return false
}

func (s *State) SetTargetMoveDelayMs(delay int) {
	s.TargetMoveDelayMs = sanitizeTargetMoveDelayMs(delay)
}

func (s *State) TargetRules() []AssistRule {
	return nil
}

func (s *State) SetTargetRules(rules []AssistRule) {
	s.targetRules.Store([]AssistRule{})
}

func (s *State) KeepAwakeOS() bool {
	return false
}

func (s *State) SetKeepAwakeOS(enabled bool) {
	s.keepAwakeOS.Store(false)
}

func (s *State) ToggleKeepAwakeOS() bool {
	s.SetKeepAwakeOS(false)
	return false
}

func (s *State) Jiggle1px() bool {
	return s.jiggle1px.Load()
}

func (s *State) SetJiggle1px(enabled bool) {
	s.jiggle1px.Store(enabled)
}

func (s *State) ToggleJiggle1px() bool {
	return toggleBool(&s.jiggle1px)
}

func (s *State) EdgeWarpEnabled() bool {
	return s.edgeWarpEnabled.Load()
}

func (s *State) SetEdgeWarpEnabled(enabled bool) {
	s.edgeWarpEnabled.Store(enabled)
}

func (s *State) ToggleEdgeWarp() bool {
	return toggleBool(&s.edgeWarpEnabled)
}

func (s *State) IdleActionsEnabled() bool {
	return s.idleActions.Load()
}

func (s *State) SetIdleActionsEnabled(enabled bool) {
	s.idleActions.Store(enabled)
}

func (s *State) ToggleIdleActions() bool {
	return toggleBool(&s.idleActions)
}

func (s *State) UpdateCheckEnabled() bool {
	return s.updateCheck.Load()
}

func (s *State) SetUpdateCheckEnabled(enabled bool) {
	s.updateCheck.Store(enabled)
}

func (s *State) ToggleUpdateCheck() bool {
	return toggleBool(&s.updateCheck)
}

func (s *State) SpriteSizePx() int {
	return int(s.spriteSize.Load())
}

func (s *State) SetSpriteSizePx(size int) {
	if size <= 0 {
		size = DefaultSettings().SpriteSizePx
	}
	if size < 16 {
		size = 16
	}
	if size > 192 {
		size = 192
	}
	s.spriteSize.Store(int32(size))
}

func (s *State) SetJiggleIdleSec(seconds int) {
	if seconds <= 0 {
		seconds = DefaultSettings().JiggleIdleSec
	}
	s.JiggleIdleSec = clampInt(seconds, 15, 300)
}

func (s *State) SetIdleActionSec(seconds int) {
	if seconds <= 0 {
		seconds = DefaultSettings().IdleActionSec
	}
	s.IdleActionSec = clampInt(seconds, 10, 180)
}

func (s *State) SetCompanionOffsets(x, y int) {
	s.CompanionOffsetX, s.CompanionOffsetY = sanitizeCompanionOffsets(x, y)
}

func (s *State) SetUpdateRepo(repo string) {
	repo = strings.TrimSpace(repo)
	if repo == "" {
		repo = DefaultSettings().UpdateRepo
	}
	s.UpdateRepo = repo
}

func (s *State) ShowPetName() bool {
	return s.showPetName.Load()
}

func (s *State) SetShowPetName(enabled bool) {
	s.showPetName.Store(enabled)
}

func (s *State) ToggleShowPetName() bool {
	return toggleBool(&s.showPetName)
}

func (s *State) SetIdleAction(action string, duration time.Duration) {
	s.idleAction.Store(action)
	s.actionUntil.Store(time.Now().Add(duration).UnixMilli())
}

func (s *State) CurrentIdleAction(now time.Time) string {
	until := s.actionUntil.Load()
	if until == 0 || now.UnixMilli() > until {
		return ""
	}
	action, _ := s.idleAction.Load().(string)
	return action
}

func (s *State) SettingsSnapshot() Settings {
	jiggle := JiggleOff
	if s.Jiggle1px() {
		jiggle = Jiggle1px
	}
	return Settings{
		PetName:            s.Name(),
		AssistantEnabled:   s.AssistantEnabled(),
		LaunchAtLogin:      s.LaunchAtLogin(),
		QuickKeysEnabled:   s.QuickKeysEnabled(),
		QuickKeys:          s.QuickKeys(),
		QuickScrollEnabled: s.QuickScrollEnabled(),
		QuickScrollLines:   SanitizeQuickScrollLines(s.QuickScrollLines),
		TargetAssist:       s.TargetAssistEnabled(),
		TargetMove:         s.TargetMoveEnabled(),
		TargetReturn:       s.TargetReturnEnabled(),
		TargetMoveDelayMs:  sanitizeTargetMoveDelayMs(s.TargetMoveDelayMs),
		TargetRules:        s.TargetRules(),
		KeepAwakeMode:      KeepAwakeOff,
		JiggleMode:         jiggle,
		JiggleIdleSec:      s.JiggleIdleSec,
		SpriteSizePx:       s.SpriteSizePx(),
		ShowPetName:        s.ShowPetName(),
		CoatColor:          s.CoatColor(),
		CompanionOffsetX:   s.CompanionOffsetX,
		CompanionOffsetY:   s.CompanionOffsetY,
		EdgeWarpEnabled:    s.EdgeWarpEnabled(),
		IdleActions:        s.IdleActionsEnabled(),
		IdleActionSec:      s.IdleActionSec,
		UpdateCheck:        s.UpdateCheckEnabled(),
		UpdateRepo:         s.UpdateRepo,
	}
}

func (s *State) StatusText() string {
	prefix := "MofuMouse: " + s.Name()
	switch {
	case s.Jiggle1px():
		return prefix + " / 1px移動中"
	case s.AssistantEnabled():
		return prefix + " / 見守り中"
	default:
		return prefix + " / お休み中"
	}
}

func sanitizeCoatColor(coat CoatColor) CoatColor {
	switch CoatColor(strings.ToLower(strings.TrimSpace(string(coat)))) {
	case CoatAgouti, CoatGray, CoatDark, CoatCream, CoatWhite, CoatPied:
		return CoatColor(strings.ToLower(strings.TrimSpace(string(coat))))
	default:
		return CoatAgouti
	}
}

func sanitizeTargetMoveDelayMs(delay int) int {
	if delay <= 0 {
		return DefaultSettings().TargetMoveDelayMs
	}
	return clampInt(delay, 200, 5000)
}

func sanitizePetName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return DefaultSettings().PetName
	}
	runes := []rune(name)
	if len(runes) > 16 {
		runes = runes[:16]
	}
	return string(runes)
}

func toggleBool(v *atomic.Bool) bool {
	for {
		old := v.Load()
		if v.CompareAndSwap(old, !old) {
			return !old
		}
	}
}
