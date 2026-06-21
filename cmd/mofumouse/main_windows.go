//go:build windows

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"fyne.io/systray"

	"mofumouse/internal/core"
	"mofumouse/internal/update"
	"mofumouse/internal/winapp"
)

var appVersion = "v0.1.0-dev"

func main() {
	winapp.EnableDPIAwareness()

	smoke := flag.Bool("smoke", false, "run a short overlay smoke test and exit")
	openSettings := flag.Bool("open-settings", false, "open the settings window on startup")
	legacySettings := flag.Bool("legacy-settings", false, "force the legacy Win32 settings window")
	configFlag := flag.String("config", "", "settings file path")
	installUpdate := flag.String("install-update", "", "install a downloaded update asset and exit")
	installTarget := flag.String("install-target", "", "current executable path to replace when using --install-update")
	installBackupDir := flag.String("install-backup-dir", "", "rollback backup directory for --install-update")
	rollbackBackup := flag.String("rollback-update", "", "restore the target executable from this backup path and exit")
	waitPID := flag.Int("wait-pid", 0, "wait for this process id to exit before installing an update")
	restartAfterInstall := flag.Bool("restart-after-install", false, "restart the installed executable after --install-update")
	restartConfig := flag.String("restart-config", "", "config path to pass when restarting after --install-update")
	flag.Parse()

	if *installUpdate != "" {
		target := *installTarget
		if target == "" {
			target, _ = os.Executable()
		}
		if *waitPID > 0 && !winapp.WaitForProcessExit(*waitPID, 45*time.Second) {
			fmt.Fprintln(os.Stderr, "timed out waiting for process to exit:", *waitPID)
			os.Exit(1)
		}
		result, err := update.InstallDownloadedAsset(*installUpdate, target, *installBackupDir)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println("installed update:", result.InstalledPath)
		fmt.Println("rollback backup:", result.BackupPath)
		if *restartAfterInstall {
			args := []string{}
			if *restartConfig != "" {
				args = append(args, "-config", *restartConfig)
			}
			if err := exec.Command(target, args...).Start(); err != nil {
				fmt.Fprintln(os.Stderr, "restart failed:", err)
				os.Exit(1)
			}
			fmt.Println("restarted:", target)
		}
		return
	}
	if *rollbackBackup != "" {
		target := *installTarget
		if target == "" {
			target, _ = os.Executable()
		}
		if err := update.RollbackInstall(target, *rollbackBackup); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println("rolled back update:", target)
		return
	}

	configPath := *configFlag
	if configPath == "" {
		var err error
		configPath, err = core.DefaultConfigPath()
		if err != nil {
			configPath = "settings.json"
		}
	}
	settings, err := core.LoadSettingsFile(configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		settings = core.DefaultSettings()
	}
	state := core.NewState(settings)

	if *smoke {
		if err := runSmoke(state); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	app := newApp(state, configPath)
	app.openSettingsOnStart = *openSettings
	app.forceLegacySettings = *legacySettings
	systray.Run(app.onReady, app.onExit)
}

type app struct {
	state               *core.State
	configPath          string
	cancel              context.CancelFunc
	wg                  sync.WaitGroup
	settingsOpen        atomic.Bool
	openSettingsOnStart bool
	forceLegacySettings bool
	pendingUpdateMu     sync.Mutex
	pendingUpdate       *pendingUpdateInfo

	menuAssistant   *systray.MenuItem
	menuLaunch      *systray.MenuItem
	menuJiggle      *systray.MenuItem
	menuIdle        *systray.MenuItem
	menuEdgeWarp    *systray.MenuItem
	menuUpdate      *systray.MenuItem
	menuUpdateCheck *systray.MenuItem
	menuUpdateApply *systray.MenuItem
	menuUpdateOpen  *systray.MenuItem
	menuSettings    *systray.MenuItem
	menuReload      *systray.MenuItem
	menuName        *systray.MenuItem
	menuCoatAgouti  *systray.MenuItem
	menuCoatGray    *systray.MenuItem
	menuCoatDark    *systray.MenuItem
	menuCoatCream   *systray.MenuItem
	menuCoatWhite   *systray.MenuItem
	menuCoatPied    *systray.MenuItem
	menuDesktop     *systray.MenuItem
	menuStart       *systray.MenuItem
	menuBack        *systray.MenuItem
	menuForward     *systray.MenuItem
	menuQuickKeys   *systray.MenuItem
	menuQuickOn     *systray.MenuItem
	menuQuickFind   *systray.MenuItem
	menuQuickCopy   *systray.MenuItem
	menuQuickAll    *systray.MenuItem
	menuQuickTab    *systray.MenuItem
	menuQuickScroll *systray.MenuItem
	menuScrollOn    *systray.MenuItem
	menuScrollUp    *systray.MenuItem
	menuScrollDown  *systray.MenuItem
	menuQuit        *systray.MenuItem
}

type pendingUpdateInfo struct {
	Version   string
	AssetName string
	Path      string
}

func newApp(state *core.State, configPath string) *app {
	return &app{state: state, configPath: configPath}
}

func (a *app) onReady() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	a.cancel = cancel

	systray.SetIcon(winapp.TrayIconBytes(true))
	systray.SetTitle("MofuMouse")
	systray.SetTooltip(a.state.StatusText())

	a.menuName = systray.AddMenuItem("デグーの名前: "+a.state.Name(), "設定画面で名前を変更できます")
	a.menuName.Disable()
	a.menuAssistant = systray.AddMenuItemCheckbox("デグーをカーソル横に表示", "マウスカーソルのすぐ横でデグーが見守ります", a.state.AssistantEnabled())
	a.menuLaunch = systray.AddMenuItemCheckbox("Windows起動時にMofuMouseを始める", "サインインしたら自動でMofuMouseを起動します", a.state.LaunchAtLogin())
	a.menuJiggle = systray.AddMenuItemCheckbox("スクリーンセーバー防止の1px移動", "触っていない時だけマウスを1pxだけ戻して動かします", a.state.Jiggle1px())
	a.menuIdle = systray.AddMenuItemCheckbox("止まっている時にしぐさを出す", "かじる、毛づくろい、ぺちぺちなどの動きを出します", a.state.IdleActionsEnabled())
	a.menuEdgeWarp = systray.AddMenuItemCheckbox("画面端で反対側へ移動", "カーソルが画面端に触れた時、反対側へ移動します", a.state.EdgeWarpEnabled())
	systray.AddSeparator()
	a.menuDesktop = systray.AddMenuItem("デスクトップ表示", "Win+D を送ります")
	a.menuStart = systray.AddMenuItem("スタートメニュー", "Windows キーを送ります")
	a.menuBack = systray.AddMenuItem("ブラウザ戻る", "Alt+Left を送ります")
	a.menuForward = systray.AddMenuItem("ブラウザ進む", "Alt+Right を送ります")
	a.menuQuickKeys = systray.AddMenuItem("クイックキー", "安全なショートカットだけを実行します")
	a.menuQuickOn = a.menuQuickKeys.AddSubMenuItemCheckbox("クイックキーを使う", "固定allowlistのショートカットだけを有効にします", a.state.QuickKeysEnabled())
	a.menuQuickKeys.AddSeparator()
	a.menuQuickFind = a.menuQuickKeys.AddSubMenuItem(quickKeyMenuTitle(core.QuickKeyFind), "検索を開きます")
	a.menuQuickCopy = a.menuQuickKeys.AddSubMenuItem(quickKeyMenuTitle(core.QuickKeyCopy), "選択中の内容をコピーします")
	a.menuQuickAll = a.menuQuickKeys.AddSubMenuItem(quickKeyMenuTitle(core.QuickKeySelectAll), "現在の入力欄や文書を全選択します")
	a.menuQuickTab = a.menuQuickKeys.AddSubMenuItem(quickKeyMenuTitle(core.QuickKeyNewTab), "対応アプリで新しいタブを開きます")
	a.menuQuickScroll = systray.AddMenuItem("クイックスクロール", "今のウィンドウを少しだけ上下にスクロールします")
	a.menuScrollOn = a.menuQuickScroll.AddSubMenuItemCheckbox("クイックスクロールを使う", "トレイから少しだけ上下スクロールできるようにします", a.state.QuickScrollEnabled())
	a.menuQuickScroll.AddSeparator()
	a.menuScrollUp = a.menuQuickScroll.AddSubMenuItem("上へ少しスクロール", "現在のウィンドウを上方向へ少しスクロールします")
	a.menuScrollDown = a.menuQuickScroll.AddSubMenuItem("下へ少しスクロール", "現在のウィンドウを下方向へ少しスクロールします")
	systray.AddSeparator()
	a.menuUpdate = systray.AddMenuItem("アップデート", "更新の確認、適用、保存場所の表示を行います")
	a.menuUpdateCheck = a.menuUpdate.AddSubMenuItem("更新を確認/ダウンロード", "新しいバージョンがあれば、このPC向けのファイルを保存します")
	a.menuUpdateApply = a.menuUpdate.AddSubMenuItem("ダウンロード済み更新を適用して再起動", "確認後にMofuMouseを終了し、更新してから再起動します")
	a.menuUpdateOpen = a.menuUpdate.AddSubMenuItem("ダウンロード済み更新を表示", "保存済み更新ファイルをエクスプローラーで表示します")
	a.setPendingUpdate(nil)
	a.menuSettings = systray.AddMenuItem("設定を開く", "名前やサイズなどを設定画面で編集します")
	a.menuReload = systray.AddMenuItem("設定を再読込", "settings.json を読み直します")
	systray.AddSeparator()
	a.menuCoatAgouti = systray.AddMenuItemCheckbox("毛色: ノーマル（茶色）", "茶色のデグーにします", a.state.CoatColor() == core.CoatAgouti)
	a.menuCoatGray = systray.AddMenuItemCheckbox("毛色: グレー", "グレーのデグーにします", a.state.CoatColor() == core.CoatGray)
	a.menuCoatDark = systray.AddMenuItemCheckbox("毛色: ダークブラウン", "濃い茶色のデグーにします", a.state.CoatColor() == core.CoatDark)
	a.menuCoatCream = systray.AddMenuItemCheckbox("毛色: クリーム", "明るいクリーム色のデグーにします", a.state.CoatColor() == core.CoatCream)
	a.menuCoatWhite = systray.AddMenuItemCheckbox("毛色: ホワイト", "白いデグーにします", a.state.CoatColor() == core.CoatWhite)
	a.menuCoatPied = systray.AddMenuItemCheckbox("毛色: ぶち模様", "ぶち模様のデグーにします", a.state.CoatColor() == core.CoatPied)
	systray.AddSeparator()
	a.menuQuit = systray.AddMenuItem("終了", "MofuMouse を終了します")

	a.wg.Add(5)
	go func() {
		defer a.wg.Done()
		if err := winapp.NewOverlay(a.state).Run(ctx); err != nil && ctx.Err() == nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}()
	go func() {
		defer a.wg.Done()
		winapp.RunJiggler(ctx, a.state)
	}()
	go func() {
		defer a.wg.Done()
		winapp.RunIdleActions(ctx, a.state)
	}()
	go func() {
		defer a.wg.Done()
		winapp.RunEdgeWarp(ctx, a.state)
	}()
	go func() {
		defer a.wg.Done()
		a.runSettingsWatcher(ctx)
	}()

	go a.handleMenu(ctx)
	if err := a.applyLaunchAtLogin(); err != nil && a.state.LaunchAtLogin() {
		systray.SetTooltip("自動起動の設定に失敗: " + err.Error())
	}
	if a.state.UpdateCheckEnabled() {
		go a.checkUpdates(false)
	}
	if a.openSettingsOnStart {
		a.openSettingsWindow(ctx)
	}
}

func (a *app) onExit() {
	if a.cancel != nil {
		a.cancel()
	}
	a.wg.Wait()
}

func (a *app) handleMenu(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			systray.Quit()
			return
		case <-a.menuAssistant.ClickedCh:
			enabled := a.state.ToggleAssistant()
			a.setChecked(a.menuAssistant, enabled)
			_ = a.persist()
			a.refreshTray()
		case <-a.menuLaunch.ClickedCh:
			enabled := a.state.ToggleLaunchAtLogin()
			if err := a.applyLaunchAtLogin(); err != nil {
				a.state.SetLaunchAtLogin(!enabled)
				a.setChecked(a.menuLaunch, !enabled)
				systray.SetTooltip("自動起動の設定に失敗: " + err.Error())
				continue
			}
			a.setChecked(a.menuLaunch, enabled)
			_ = a.persist()
			a.refreshTray()
		case <-a.menuJiggle.ClickedCh:
			enabled := a.state.ToggleJiggle1px()
			a.setChecked(a.menuJiggle, enabled)
			_ = a.persist()
			a.refreshTray()
		case <-a.menuIdle.ClickedCh:
			enabled := a.state.ToggleIdleActions()
			a.setChecked(a.menuIdle, enabled)
			_ = a.persist()
			a.refreshTray()
		case <-a.menuEdgeWarp.ClickedCh:
			enabled := a.state.ToggleEdgeWarp()
			a.setChecked(a.menuEdgeWarp, enabled)
			_ = a.persist()
			a.refreshTray()
		case <-a.menuUpdateCheck.ClickedCh:
			go a.checkUpdates(true)
		case <-a.menuUpdateApply.ClickedCh:
			go a.applyPendingUpdate()
		case <-a.menuUpdateOpen.ClickedCh:
			a.openPendingUpdate()
		case <-a.menuSettings.ClickedCh:
			a.openSettingsWindow(ctx)
		case <-a.menuReload.ClickedCh:
			a.reloadSettings()
		case <-a.menuCoatAgouti.ClickedCh:
			a.setCoatColor(core.CoatAgouti)
		case <-a.menuCoatGray.ClickedCh:
			a.setCoatColor(core.CoatGray)
		case <-a.menuCoatDark.ClickedCh:
			a.setCoatColor(core.CoatDark)
		case <-a.menuCoatCream.ClickedCh:
			a.setCoatColor(core.CoatCream)
		case <-a.menuCoatWhite.ClickedCh:
			a.setCoatColor(core.CoatWhite)
		case <-a.menuCoatPied.ClickedCh:
			a.setCoatColor(core.CoatPied)
		case <-a.menuDesktop.ClickedCh:
			winapp.ShowDesktop()
		case <-a.menuStart.ClickedCh:
			winapp.OpenStartMenu()
		case <-a.menuBack.ClickedCh:
			winapp.BrowserBack()
		case <-a.menuForward.ClickedCh:
			winapp.BrowserForward()
		case <-a.menuQuickOn.ClickedCh:
			enabled := a.state.ToggleQuickKeysEnabled()
			a.setChecked(a.menuQuickOn, enabled)
			a.refreshQuickKeyMenu()
			_ = a.persist()
		case <-a.menuQuickFind.ClickedCh:
			a.runQuickKey(core.QuickKeyFind)
		case <-a.menuQuickCopy.ClickedCh:
			a.runQuickKey(core.QuickKeyCopy)
		case <-a.menuQuickAll.ClickedCh:
			a.runQuickKey(core.QuickKeySelectAll)
		case <-a.menuQuickTab.ClickedCh:
			a.runQuickKey(core.QuickKeyNewTab)
		case <-a.menuScrollOn.ClickedCh:
			enabled := a.state.ToggleQuickScrollEnabled()
			a.setChecked(a.menuScrollOn, enabled)
			a.refreshQuickScrollMenu()
			_ = a.persist()
		case <-a.menuScrollUp.ClickedCh:
			a.runQuickScroll(core.QuickScrollUp)
		case <-a.menuScrollDown.ClickedCh:
			a.runQuickScroll(core.QuickScrollDown)
		case <-a.menuQuit.ClickedCh:
			systray.Quit()
			return
		}
	}
}

func (a *app) setChecked(item *systray.MenuItem, checked bool) {
	if checked {
		item.Check()
		return
	}
	item.Uncheck()
}

func (a *app) refreshTray() {
	systray.SetIcon(winapp.TrayIconBytes(a.state.AssistantEnabled()))
	systray.SetTooltip(a.state.StatusText())
	if a.menuName != nil {
		a.menuName.SetTitle("デグーの名前: " + a.state.Name())
	}
	a.refreshMenuChecks()
}

func (a *app) refreshMenuChecks() {
	if a.menuAssistant != nil {
		a.setChecked(a.menuAssistant, a.state.AssistantEnabled())
	}
	if a.menuLaunch != nil {
		a.setChecked(a.menuLaunch, a.state.LaunchAtLogin())
	}
	if a.menuJiggle != nil {
		a.setChecked(a.menuJiggle, a.state.Jiggle1px())
	}
	if a.menuIdle != nil {
		a.setChecked(a.menuIdle, a.state.IdleActionsEnabled())
	}
	if a.menuEdgeWarp != nil {
		a.setChecked(a.menuEdgeWarp, a.state.EdgeWarpEnabled())
	}
	a.refreshQuickKeyMenu()
	a.refreshQuickScrollMenu()
	a.refreshCoatChecks()
}

func (a *app) refreshQuickKeyMenu() {
	if a.menuQuickKeys == nil {
		return
	}
	enabled := a.state.QuickKeysEnabled()
	a.setChecked(a.menuQuickOn, enabled)
	a.setQuickKeyItemState(a.menuQuickFind, core.QuickKeyFind)
	a.setQuickKeyItemState(a.menuQuickCopy, core.QuickKeyCopy)
	a.setQuickKeyItemState(a.menuQuickAll, core.QuickKeySelectAll)
	a.setQuickKeyItemState(a.menuQuickTab, core.QuickKeyNewTab)
}

func (a *app) setQuickKeyItemState(item *systray.MenuItem, key core.QuickKeyID) {
	if item == nil {
		return
	}
	if a.state.QuickKeyAllowed(key) {
		item.Enable()
		return
	}
	item.Disable()
}

func (a *app) refreshQuickScrollMenu() {
	if a.menuQuickScroll == nil {
		return
	}
	enabled := a.state.QuickScrollEnabled()
	a.setChecked(a.menuScrollOn, enabled)
	a.setScrollItemState(a.menuScrollUp, core.QuickScrollUp)
	a.setScrollItemState(a.menuScrollDown, core.QuickScrollDown)
}

func (a *app) setScrollItemState(item *systray.MenuItem, direction core.QuickScrollDirection) {
	if item == nil {
		return
	}
	if a.state.QuickScrollAllowed(direction) {
		item.Enable()
		return
	}
	item.Disable()
}

func (a *app) runQuickKey(key core.QuickKeyID) {
	if !a.state.QuickKeyAllowed(key) {
		systray.SetTooltip("このクイックキーは無効です")
		return
	}
	if !winapp.RunQuickKey(key) {
		systray.SetTooltip("このクイックキーは実行できません")
		return
	}
	systray.SetTooltip("クイックキー: " + core.QuickKeyLabel(key) + " (" + core.QuickKeyShortcut(key) + ")")
}

func (a *app) runQuickScroll(direction core.QuickScrollDirection) {
	if !a.state.QuickScrollAllowed(direction) {
		systray.SetTooltip("クイックスクロールは無効です")
		return
	}
	if !winapp.RunQuickScroll(direction, a.state.QuickScrollLines) {
		systray.SetTooltip("クイックスクロールを実行できません")
		return
	}
	a.state.SetIdleAction("sniff", 900*time.Millisecond)
	systray.SetTooltip(core.QuickScrollLabel(direction))
}

func quickKeyMenuTitle(key core.QuickKeyID) string {
	return core.QuickKeyLabel(key) + " (" + core.QuickKeyShortcut(key) + ")"
}

func (a *app) setCoatColor(coat core.CoatColor) {
	a.state.SetCoatColor(coat)
	_ = a.persist()
	a.refreshTray()
}

func (a *app) refreshCoatChecks() {
	if a.menuCoatAgouti == nil {
		return
	}
	coat := a.state.CoatColor()
	a.setChecked(a.menuCoatAgouti, coat == core.CoatAgouti)
	a.setChecked(a.menuCoatGray, coat == core.CoatGray)
	a.setChecked(a.menuCoatDark, coat == core.CoatDark)
	a.setChecked(a.menuCoatCream, coat == core.CoatCream)
	a.setChecked(a.menuCoatWhite, coat == core.CoatWhite)
	a.setChecked(a.menuCoatPied, coat == core.CoatPied)
}

func (a *app) persist() error {
	return core.SaveSettingsFile(a.configPath, a.state.SettingsSnapshot())
}

func (a *app) applyLaunchAtLogin() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	return winapp.SetLaunchAtLogin(a.state.LaunchAtLogin(), exe)
}

func (a *app) openSettingsWindow(ctx context.Context) {
	if !a.settingsOpen.CompareAndSwap(false, true) {
		systray.SetTooltip("設定画面はすでに開いています")
		return
	}
	if !a.forceLegacySettings && a.openModernSettingsWindow(ctx) {
		return
	}
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		defer a.settingsOpen.Store(false)
		err := winapp.RunSettingsWindow(ctx, a.state, winapp.SettingsWindowHooks{
			Save: a.persist,
			Refresh: func() {
				if err := a.applyLaunchAtLogin(); err != nil {
					systray.SetTooltip("自動起動の設定に失敗: " + err.Error())
				}
				a.refreshTray()
			},
			OpenConfig: func() {
				_ = exec.Command("notepad.exe", a.configPath).Start()
			},
			CheckUpdates: func() {
				go a.checkUpdates(true)
			},
		})
		if err != nil && ctx.Err() == nil {
			systray.SetTooltip("設定画面を開けません: " + err.Error())
		}
	}()
}

func (a *app) openModernSettingsWindow(ctx context.Context) bool {
	path, ok := modernSettingsExecutable()
	if !ok {
		return false
	}
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		defer a.settingsOpen.Store(false)

		cmd := exec.CommandContext(ctx, path, "-config", a.configPath)
		if err := cmd.Run(); err != nil && ctx.Err() == nil {
			systray.SetTooltip("新しい設定画面を開けません: " + err.Error())
			return
		}
		a.reloadSettings()
	}()
	return true
}

func modernSettingsExecutable() (string, bool) {
	exe, err := os.Executable()
	if err != nil {
		return "", false
	}
	dir := filepath.Dir(exe)
	suffix := "x64"
	if runtime.GOARCH == "386" {
		suffix = "x86"
	}
	candidates := []string{
		filepath.Join(dir, "mofumouse-control-"+suffix+".exe"),
		filepath.Join(dir, "mofumouse-control.exe"),
		filepath.Join(dir, "mofumouse-control-x64.exe"),
	}
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, true
		}
	}
	return "", false
}

func (a *app) runSettingsWatcher(ctx context.Context) {
	lastMod, lastSize, _ := configSignature(a.configPath)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			mod, size, ok := configSignature(a.configPath)
			if !ok || (mod == lastMod && size == lastSize) {
				continue
			}
			settings, err := core.LoadSettingsFile(a.configPath)
			if err != nil {
				continue
			}
			lastMod, lastSize = mod, size
			a.applySettings(settings)
		}
	}
}

func configSignature(path string) (int64, int64, bool) {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return 0, 0, false
	}
	return info.ModTime().UnixNano(), info.Size(), true
}

func (a *app) reloadSettings() {
	settings, err := core.LoadSettingsFile(a.configPath)
	if err != nil {
		systray.SetTooltip("設定の再読込に失敗: " + err.Error())
		return
	}
	a.applySettings(settings)
	if err := a.applyLaunchAtLogin(); err != nil {
		systray.SetTooltip("自動起動の設定に失敗: " + err.Error())
	}
	a.refreshTray()
}

func (a *app) applySettings(settings core.Settings) {
	launchChanged := settings.LaunchAtLogin != a.state.LaunchAtLogin()
	a.state.ApplySettings(settings)
	if !launchChanged {
		a.refreshTray()
		return
	}
	if err := a.applyLaunchAtLogin(); err != nil {
		systray.SetTooltip("自動起動の設定に失敗: " + err.Error())
	}
	a.refreshTray()
}

func (a *app) setPendingUpdate(info *pendingUpdateInfo) {
	a.pendingUpdateMu.Lock()
	if info == nil || info.Path == "" {
		a.pendingUpdate = nil
	} else {
		copy := *info
		a.pendingUpdate = &copy
	}
	a.pendingUpdateMu.Unlock()

	if a.menuUpdateApply == nil || a.menuUpdateOpen == nil {
		return
	}
	if info == nil || info.Path == "" {
		a.menuUpdateApply.Disable()
		a.menuUpdateOpen.Disable()
		return
	}
	a.menuUpdateApply.Enable()
	a.menuUpdateOpen.Enable()
}

func (a *app) pendingUpdateSnapshot() (pendingUpdateInfo, bool) {
	a.pendingUpdateMu.Lock()
	defer a.pendingUpdateMu.Unlock()
	if a.pendingUpdate == nil || a.pendingUpdate.Path == "" {
		return pendingUpdateInfo{}, false
	}
	return *a.pendingUpdate, true
}

func (a *app) openPendingUpdate() {
	pending, ok := a.pendingUpdateSnapshot()
	if !ok {
		systray.SetTooltip("ダウンロード済み更新はまだありません")
		return
	}
	if _, err := os.Stat(pending.Path); err != nil {
		systray.SetTooltip("更新ファイルが見つかりません: " + err.Error())
		a.setPendingUpdate(nil)
		return
	}
	_ = exec.Command("explorer.exe", "/select,"+pending.Path).Start()
}

func (a *app) applyPendingUpdate() {
	pending, ok := a.pendingUpdateSnapshot()
	if !ok {
		systray.SetTooltip("先にアップデートをダウンロードしてください")
		return
	}
	if _, err := os.Stat(pending.Path); err != nil {
		systray.SetTooltip("更新ファイルが見つかりません: " + err.Error())
		a.setPendingUpdate(nil)
		return
	}
	message := "MofuMouseを終了して更新を適用し、完了後に再起動します。\n\nバージョン: " + pending.Version + "\nファイル: " + pending.AssetName + "\n\n今すぐ続行しますか？"
	if !winapp.ConfirmYesNo("MofuMouse の更新", message) {
		systray.SetTooltip("更新の適用をキャンセルしました")
		return
	}

	exe, err := os.Executable()
	if err != nil {
		systray.SetTooltip("更新ヘルパーを準備できません: " + err.Error())
		return
	}
	helperDir, err := update.DefaultHelperDir()
	if err != nil {
		systray.SetTooltip("更新ヘルパーを準備できません: " + err.Error())
		return
	}
	helper, err := update.PrepareHelperExecutable(exe, helperDir)
	if err != nil {
		systray.SetTooltip("更新ヘルパーを準備できません: " + err.Error())
		return
	}

	logPath := filepath.Join(helperDir, "mofumouse-update-"+time.Now().UTC().Format("20060102-150405.000000000")+".log")
	logFile, _ := os.Create(logPath)
	if logFile != nil {
		defer logFile.Close()
	}

	args := []string{
		"--wait-pid", strconv.Itoa(os.Getpid()),
		"--install-update", pending.Path,
		"--install-target", exe,
		"--install-backup-dir", filepath.Join(helperDir, "rollback"),
		"--restart-after-install",
	}
	if a.configPath != "" {
		args = append(args, "--restart-config", a.configPath)
	}
	cmd := exec.Command(helper, args...)
	if logFile != nil {
		cmd.Stdout = logFile
		cmd.Stderr = logFile
	}
	if err := cmd.Start(); err != nil {
		systray.SetTooltip("更新ヘルパーを起動できません: " + err.Error())
		return
	}
	systray.SetTooltip("更新を適用します。MofuMouseを再起動します。")
	systray.Quit()
}

func (a *app) checkUpdates(download bool) {
	a.menuUpdate.SetTitle("アップデート確認中...")
	a.menuUpdateCheck.SetTitle("確認中...")
	result, err := update.CheckLatest(a.state.UpdateRepo, appVersion)
	if err != nil {
		a.menuUpdate.SetTitle("アップデート確認に失敗")
		a.menuUpdateCheck.SetTitle("更新を確認/ダウンロード")
		systray.SetTooltip("アップデート確認に失敗: " + err.Error())
		time.Sleep(4 * time.Second)
		a.menuUpdate.SetTitle("アップデート")
		a.refreshTray()
		return
	}
	if result.UpdateAvailable {
		a.menuUpdate.SetTitle("アップデートあり: " + result.LatestVersion)
		a.menuUpdateCheck.SetTitle("更新をダウンロード: " + result.LatestVersion)
		if result.AssetURL == "" {
			systray.SetTooltip("アップデートはありますが、このPC向けのファイルが見つかりません。")
			if download && result.ReleaseURL != "" {
				_ = exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", result.ReleaseURL).Start()
			}
			return
		}
		if !download {
			systray.SetTooltip("アップデートあり: " + result.LatestVersion + " / " + result.AssetName)
			return
		}
		a.menuUpdate.SetTitle("アップデートをダウンロード中...")
		a.menuUpdateCheck.SetTitle("ダウンロード中...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		downloaded, err := update.DownloadAsset(ctx, result.AssetURL, result.AssetName, result.LatestVersion)
		if err != nil {
			a.menuUpdate.SetTitle("ダウンロードに失敗")
			a.menuUpdateCheck.SetTitle("更新を確認/ダウンロード")
			systray.SetTooltip("アップデートのダウンロードに失敗: " + err.Error())
			time.Sleep(4 * time.Second)
			a.menuUpdate.SetTitle("アップデート")
			a.refreshTray()
			return
		}
		a.menuUpdate.SetTitle("ダウンロード済み: " + result.LatestVersion)
		a.menuUpdateCheck.SetTitle("再確認/再ダウンロード")
		a.setPendingUpdate(&pendingUpdateInfo{
			Version:   result.LatestVersion,
			AssetName: downloaded.AssetName,
			Path:      downloaded.Path,
		})
		systray.SetTooltip("アップデートを保存しました。トレイの「適用して再起動」から確認できます。")
		return
	}
	a.setPendingUpdate(nil)
	a.menuUpdate.SetTitle("最新版です")
	a.menuUpdateCheck.SetTitle("更新を確認/ダウンロード")
	time.Sleep(4 * time.Second)
	a.menuUpdate.SetTitle("アップデート")
	a.refreshTray()
}

func runSmoke(state *core.State) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- winapp.NewOverlay(state).Run(ctx)
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(7 * time.Second):
		return fmt.Errorf("smoke timed out")
	}
}
