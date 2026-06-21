import { useEffect, useMemo, useRef, useState, type ReactNode } from "react";
import {
  Badge,
  Button,
  Divider,
  Dropdown,
  Field,
  Input,
  Option,
  Slider,
  Spinner,
  Switch,
  Tab,
  TabList,
  Text,
  type TabValue
} from "@fluentui/react-components";
import {
  AnimalPawPrint24Regular,
  ArrowReset24Regular,
  ArrowSync24Regular,
  Color24Regular,
  Desktop24Regular,
  DocumentSave24Regular,
  FolderOpen24Regular,
  Home24Regular,
  Keyboard24Regular,
  PanelLeft24Regular,
  Save24Regular,
  Settings24Regular,
  WrenchSettings24Regular
} from "@fluentui/react-icons";

import { ConfigPath, GetSettings, OpenConfigFile, ResetSettings, SaveSettings } from "../wailsjs/go/main/ControlCenterApp";
import { core } from "../wailsjs/go/models";
import deguIdleAgouti from "./assets/degu_idle_agouti.png";
import deguIdleCream from "./assets/degu_idle_cream.png";
import deguIdleDark from "./assets/degu_idle_dark.png";
import deguIdleGray from "./assets/degu_idle_gray.png";
import deguIdlePied from "./assets/degu_idle_pied.png";
import deguIdleWhite from "./assets/degu_idle_white.png";
import {
  coatLabels,
  defaultSettings,
  quickKeyLabels,
  quickKeyShortcuts,
  type CoatColor,
  type JiggleMode,
  type QuickKeyID,
  type Settings
} from "./settings";

type Status = "loading" | "ready" | "saving" | "saved" | "error";

const tabs = [
  { value: "home", label: "Home", icon: <Home24Regular /> },
  { value: "companion", label: "Companion", icon: <AnimalPawPrint24Regular /> },
  { value: "jiggle", label: "Jiggle", icon: <Desktop24Regular /> },
  { value: "appearance", label: "Appearance", icon: <Color24Regular /> },
  { value: "updates", label: "Updates", icon: <ArrowSync24Regular /> },
  { value: "advanced", label: "Advanced", icon: <WrenchSettings24Regular /> }
];

const coatImages: Record<CoatColor, string> = {
  agouti: deguIdleAgouti,
  gray: deguIdleGray,
  dark: deguIdleDark,
  cream: deguIdleCream,
  white: deguIdleWhite,
  pied: deguIdlePied
};

function previewImageForCoat(coat: CoatColor) {
  return coatImages[coat] ?? deguIdleAgouti;
}

function App() {
  const [activeTab, setActiveTab] = useState<TabValue>("home");
  const [settings, setSettings] = useState<Settings>(defaultSettings);
  const [savedSettings, setSavedSettings] = useState<Settings>(defaultSettings);
  const [configPath, setConfigPath] = useState("");
  const [status, setStatus] = useState<Status>("loading");
  const saveVersion = useRef(0);
  const [message, setMessage] = useState("設定を読み込んでいます。");

  useEffect(() => {
    async function load() {
      try {
        const [loaded, path] = await Promise.all([GetSettings(), ConfigPath()]);
        const normalized = normalizeSettings(loaded);
        setSettings(normalized);
        setSavedSettings(normalized);
        setConfigPath(path);
        setStatus("ready");
        setMessage("設定を読み込みました。変更したら保存してください。");
      } catch (error) {
        setStatus("error");
        setMessage(errorMessage(error));
      }
    }
    load();
  }, []);

  const dirty = useMemo(() => JSON.stringify(settings) !== JSON.stringify(savedSettings), [settings, savedSettings]);

  useEffect(() => {
    if (!dirty) {
      return;
    }
    const version = ++saveVersion.current;
    setStatus("saving");
    setMessage("変更を自動保存しています。");

    const timer = window.setTimeout(async () => {
      try {
        const saved = await SaveSettings(core.Settings.createFrom(settings));
        if (version !== saveVersion.current) {
          return;
        }
        const normalized = normalizeSettings(saved);
        setSettings(normalized);
        setSavedSettings(normalized);
        setStatus("saved");
        setMessage("変更を保存しました。MofuMouseに反映されます。");
      } catch (error) {
        if (version !== saveVersion.current) {
          return;
        }
        setStatus("error");
        setMessage(errorMessage(error));
      }
    }, 350);

    return () => window.clearTimeout(timer);
  }, [dirty, settings]);

  function patch(next: Partial<Settings>) {
    setSettings((current) => ({ ...current, ...next }));
    if (status === "saved") {
      setStatus("ready");
      setMessage("未保存の変更があります。");
    }
  }

  async function save() {
    saveVersion.current += 1;
    setStatus("saving");
    setMessage("設定を保存しています。");
    try {
      const saved = await SaveSettings(core.Settings.createFrom(settings));
      const normalized = normalizeSettings(saved);
      setSettings(normalized);
      setSavedSettings(normalized);
      setStatus("saved");
      setMessage("保存しました。常駐中の本体には、閉じた後に反映されます。");
    } catch (error) {
      setStatus("error");
      setMessage(errorMessage(error));
    }
  }

  async function reset() {
    saveVersion.current += 1;
    setStatus("saving");
    setMessage("初期設定に戻しています。");
    try {
      const resetSettings = await ResetSettings();
      const normalized = normalizeSettings(resetSettings);
      setSettings(normalized);
      setSavedSettings(normalized);
      setStatus("saved");
      setMessage("初期設定へ戻しました。");
    } catch (error) {
      setStatus("error");
      setMessage(errorMessage(error));
    }
  }

  async function openConfig() {
    try {
      await OpenConfigFile();
    } catch (error) {
      setStatus("error");
      setMessage(errorMessage(error));
    }
  }

  return (
    <div className="appShell">
      <aside className="sidebar" aria-label="MofuMouse settings sections">
        <div className="brand">
          <img src={deguIdleAgouti} alt="" />
          <div>
            <strong>MofuMouse</strong>
            <span>Control Center</span>
          </div>
        </div>
        <TabList
          vertical
          selectedValue={activeTab}
          onTabSelect={(_, data) => setActiveTab(data.value)}
          className="navTabs"
        >
          {tabs.map((tab) => (
            <Tab key={tab.value} value={tab.value} icon={tab.icon}>
              {tab.label}
            </Tab>
          ))}
        </TabList>
      </aside>

      <main className="content">
        <header className="topBar">
          <div>
            <Text className="eyebrow">設定ファイル</Text>
            <h1>{headingFor(activeTab)}</h1>
            <Text className="pathText">{configPath || "読み込み中"}</Text>
          </div>
          <div className="actions">
            <Button icon={<FolderOpen24Regular />} onClick={openConfig}>
              設定ファイル
            </Button>
            <Button icon={<ArrowReset24Regular />} onClick={reset}>
              初期化
            </Button>
            <Button appearance="primary" icon={<Save24Regular />} disabled={!dirty || status === "saving"} onClick={save}>
              保存
            </Button>
          </div>
        </header>

        <LiveStatusBar status={status} dirty={dirty} message={message} />

        <div className="tabSurface">
          {activeTab === "home" && <HomeTab settings={settings} patch={patch} />}
          {activeTab === "companion" && <CompanionTab settings={settings} patch={patch} />}
          {activeTab === "jiggle" && <JiggleTab settings={settings} patch={patch} />}
          {activeTab === "appearance" && <AppearanceTab settings={settings} patch={patch} />}
          {activeTab === "updates" && <UpdatesTab settings={settings} patch={patch} />}
          {activeTab === "advanced" && <AdvancedTab settings={settings} patch={patch} />}
        </div>
      </main>
    </div>
  );
}

function HomeTab({ settings, patch }: TabProps) {
  return (
    <div className="twoColumn">
      <section className="sectionBlock heroPreview">
        <div>
          <Badge appearance="filled" color={settings.AssistantEnabled ? "success" : "subtle"}>
            {settings.AssistantEnabled ? "見守り中" : "お休み中"}
          </Badge>
          <h2>カーソルの横で手伝うデグー</h2>
          <p>表示、歩いてついてくる動き、1px移動をここから切り替えます。</p>
        </div>
        <div className="previewStage">
          <div className="cursorShape" />
          <img src={previewImageForCoat(settings.CoatColor)} alt="MofuMouse degu preview" />
        </div>
      </section>
      <section className="sectionBlock stack">
        <SwitchRow
          icon={<Desktop24Regular />}
          title="カーソル横に表示"
          detail="マウスのすぐ横にデグーを表示します。"
          checked={settings.AssistantEnabled}
          onChange={(checked) => patch({ AssistantEnabled: checked })}
        />
        <SwitchRow
          icon={<Desktop24Regular />}
          title="Windows起動時に始める"
          detail="サインインしたら自動でMofuMouseを起動します。"
          checked={settings.LaunchAtLogin}
          onChange={(checked) => patch({ LaunchAtLogin: checked })}
        />
        <SwitchRow
          icon={<Desktop24Regular />}
          title="1pxだけ動かす"
          detail="しばらく触っていない時だけ、マウスを1pxだけ戻して動かします。"
          checked={settings.JiggleMode === "1px"}
          onChange={(checked) => patch({ JiggleMode: checked ? "1px" : "off" })}
        />
      </section>
    </div>
  );
}

function CompanionTab({ settings, patch }: TabProps) {
  return (
    <div className="panelGrid">
      <section className="sectionBlock stack">
        <SectionHeader title="動き" description="カーソル横の基本動作を調整します。" />
        <SwitchRow
          icon={<Desktop24Regular />}
          title="カーソル横にデグーを表示"
          detail="オフにするとoverlayを隠します。"
          checked={settings.AssistantEnabled}
          onChange={(checked) => patch({ AssistantEnabled: checked })}
        />
        <SwitchRow
          icon={<AnimalPawPrint24Regular />}
          title="止まっている時にしぐさを出す"
          detail="かじる、毛づくろい、ぺちぺちなどの待機アクションを出します。"
          checked={settings.IdleActions}
          onChange={(checked) => patch({ IdleActions: checked })}
        />
        <Field label={`待機アクションまで ${settings.IdleActionSec} 秒`}>
          <Slider min={10} max={180} step={5} value={settings.IdleActionSec} onChange={(_, data) => patch({ IdleActionSec: data.value })} />
        </Field>
      </section>
      <section className="sectionBlock stack">
        <SectionHeader title="位置" description="既定ではカーソルの真横に置きます。" />
        <Field label={`横位置 ${signedOffset(settings.CompanionOffsetX)}`}>
          <Slider min={-48} max={48} step={2} value={settings.CompanionOffsetX} onChange={(_, data) => patch({ CompanionOffsetX: data.value })} />
        </Field>
        <Field label={verticalOffsetLabel(settings.CompanionOffsetY)}>
          <Slider min={-48} max={48} step={2} value={settings.CompanionOffsetY} onChange={(_, data) => patch({ CompanionOffsetY: data.value })} />
        </Field>
        <SwitchRow
          icon={<PanelLeft24Regular />}
          title="画面端で反対側へ移動"
          detail="慣れてから使う設定です。"
          checked={settings.EdgeWarpEnabled}
          onChange={(checked) => patch({ EdgeWarpEnabled: checked })}
        />
      </section>
    </div>
  );
}

const quickKeyOrder: QuickKeyID[] = ["find", "copy", "select_all", "new_tab"];

function QuickKeysEditor({ settings, patch }: TabProps) {
  function toggleKey(key: QuickKeyID, checked: boolean) {
    const current = new Set(settings.QuickKeys);
    if (checked) {
      current.add(key);
    } else {
      current.delete(key);
    }
    patch({ QuickKeys: quickKeyOrder.filter((candidate) => current.has(candidate)) });
  }

  return (
    <div className="quickKeyEditor">
      <SectionHeader
        title="クイックキー"
        description="トレイから実行できる安全な固定ショートカットです。削除、送信、閉じる操作は入れません。"
      />
      <SwitchRow
        icon={<Keyboard24Regular />}
        title="クイックキーを使う"
        detail="オンにするとトレイのクイックキー項目が使えます。"
        checked={settings.QuickKeysEnabled}
        onChange={(checked) => patch({ QuickKeysEnabled: checked })}
      />
      <div className="quickKeyList">
        {quickKeyOrder.map((key) => (
          <label key={key} className="quickKeyItem">
            <Switch checked={settings.QuickKeys.includes(key)} onChange={(_, data) => toggleKey(key, Boolean(data.checked))} />
            <span>
              <strong>{quickKeyLabels[key]}</strong>
              <small>{quickKeyShortcuts[key]}</small>
            </span>
          </label>
        ))}
      </div>
    </div>
  );
}

function QuickScrollEditor({ settings, patch }: TabProps) {
  return (
    <div className="quickScrollEditor">
      <SectionHeader
        title="クイックスクロール"
        description="トレイから、今見ているウィンドウを少しだけ上下にスクロールできます。勝手に連続スクロールはしません。"
      />
      <SwitchRow
        icon={<PanelLeft24Regular />}
        title="クイックスクロールを使う"
        detail="オンにすると、トレイに上へ/下へスクロールの項目が出ます。"
        checked={settings.QuickScrollEnabled}
        onChange={(checked) => patch({ QuickScrollEnabled: checked })}
      />
      <Field label={`1回のスクロール量 ${settings.QuickScrollLines} 行`}>
        <Slider
          min={1}
          max={10}
          step={1}
          value={settings.QuickScrollLines}
          onChange={(_, data) => patch({ QuickScrollLines: data.value })}
        />
      </Field>
    </div>
  );
}

function JiggleTab({ settings, patch }: TabProps) {
  return (
    <div className="panelGrid">
      <section className="sectionBlock stack">
        <SectionHeader title="1px移動" description="PCを起こし続けたい時は、OS制御ではなく最小のマウス移動だけを使います。" />
        <SwitchRow
          icon={<Desktop24Regular />}
          title="スクリーンセーバー防止の1px移動"
          detail="触っていない時だけ1pxだけ戻して動かします。"
          checked={settings.JiggleMode === "1px"}
          onChange={(checked) => patch({ JiggleMode: checked ? "1px" : "off" })}
        />
        <Field label={`1px移動まで ${settings.JiggleIdleSec} 秒`}>
          <Slider min={15} max={300} step={5} value={settings.JiggleIdleSec} onChange={(_, data) => patch({ JiggleIdleSec: data.value })} />
        </Field>
      </section>
      <section className="sectionBlock stack">
        <SectionHeader title="補助ツール" description="トレイから呼び出す固定操作だけを残しています。" />
        <QuickKeysEditor settings={settings} patch={patch} />
        <QuickScrollEditor settings={settings} patch={patch} />
      </section>
    </div>
  );
}

function AppearanceTab({ settings, patch }: TabProps) {
  return (
    <div className="panelGrid">
      <section className="sectionBlock stack">
        <SectionHeader title="名前と表示" description="名前は常時表示しないのが既定です。" />
        <Field label="デグーの名前">
          <Input value={settings.PetName} maxLength={16} onChange={(_, data) => patch({ PetName: data.value })} />
        </Field>
        <SwitchRow
          icon={<AnimalPawPrint24Regular />}
          title="デグーの名前を画面に出す"
          detail="普段はオフ推奨です。"
          checked={settings.ShowPetName}
          onChange={(checked) => patch({ ShowPetName: checked })}
        />
        <Field label={`大きさ ${settings.SpriteSizePx}px`}>
          <Slider min={16} max={96} step={4} value={settings.SpriteSizePx} onChange={(_, data) => patch({ SpriteSizePx: data.value })} />
        </Field>
      </section>
      <section className="sectionBlock stack">
        <SectionHeader title="毛色" description="見た目だけの設定です。" />
        <Dropdown
          value={coatLabels[settings.CoatColor]}
          selectedOptions={[settings.CoatColor]}
          onOptionSelect={(_, data) => patch({ CoatColor: data.optionValue as CoatColor })}
        >
          {(Object.keys(coatLabels) as CoatColor[]).map((coat) => (
            <Option key={coat} value={coat}>
              {coatLabels[coat]}
            </Option>
          ))}
        </Dropdown>
        <div className="coatPreview">
          <img src={previewImageForCoat(settings.CoatColor)} alt="" />
          <span>{coatLabels[settings.CoatColor]}</span>
        </div>
      </section>
    </div>
  );
}

function UpdatesTab({ settings, patch }: TabProps) {
  return (
    <div className="panelGrid">
      <section className="sectionBlock stack">
        <SectionHeader title="アップデート" description="GitHub Releases の確認、ダウンロード、確認付き適用と再起動に対応しています。" />
        <SwitchRow
          icon={<ArrowSync24Regular />}
          title="起動時に更新を確認"
          detail="新しい版が見つかった時にトレイで知らせます。適用はトレイで確認してから行います。"
          checked={settings.UpdateCheck}
          onChange={(checked) => patch({ UpdateCheck: checked })}
        />
        <Field label="更新確認リポジトリ">
          <Input value={settings.UpdateRepo} onChange={(_, data) => patch({ UpdateRepo: data.value })} />
        </Field>
      </section>
      <section className="sectionBlock">
        <SectionHeader title="リリース条件" description="QAが終わるまで公開しません。" />
        <ul className="releaseList">
          <li>x86/x64 本体ビルド</li>
          <li>x86/x64 更新ファイルの選択、ダウンロード、確認付き適用</li>
          <li>歩行モーション / 1px移動 QA</li>
          <li>ImageGen アセット全ポーズ確認</li>
          <li>GitHub Release と Pages は最後に有効化</li>
        </ul>
      </section>
    </div>
  );
}

function AdvancedTab({ settings, patch }: TabProps) {
  return (
    <div className="panelGrid">
      <section className="sectionBlock stack">
        <SectionHeader title="保存値" description="通常は触らなくてよい設定です。" />
        <Field label="JiggleMode">
          <Dropdown
            value={settings.JiggleMode}
            selectedOptions={[settings.JiggleMode]}
            onOptionSelect={(_, data) => patch({ JiggleMode: data.optionValue as JiggleMode })}
          >
            <Option value="off">off</Option>
            <Option value="1px">1px</Option>
          </Dropdown>
        </Field>
      </section>
      <section className="sectionBlock">
        <SectionHeader title="安全境界" description="このアプリがやらないこと。" />
        <div className="boundaryList">
          <span>カーソルを勝手に移動しない</span>
          <span>自動クリックしない</span>
          <span>パスワードを読まない</span>
          <span>UACを操作しない</span>
        </div>
      </section>
    </div>
  );
}

type TabProps = {
  settings: Settings;
  patch: (next: Partial<Settings>) => void;
};

type SwitchRowProps = {
  icon: ReactNode;
  title: string;
  detail: string;
  checked: boolean;
  onChange: (checked: boolean) => void;
};

function SwitchRow({ icon, title, detail, checked, onChange }: SwitchRowProps) {
  return (
    <div className="switchRow">
      <div className="rowIcon">{icon}</div>
      <div>
        <strong>{title}</strong>
        <span>{detail}</span>
      </div>
      <Switch checked={checked} onChange={(_, data) => onChange(Boolean(data.checked))} />
    </div>
  );
}

function SectionHeader({ title, description }: { title: string; description: string }) {
  return (
    <div className="sectionHeader">
      <h2>{title}</h2>
      <p>{description}</p>
    </div>
  );
}

function LiveStatusBar({ status, dirty, message }: { status: Status; dirty: boolean; message: string }) {
  const badgeLabel = status === "saving" ? "反映中" : dirty ? "自動保存待ち" : status === "error" ? "確認が必要" : "保存済み";
  const badgeColor = status === "error" ? "danger" : status === "saving" || dirty ? "warning" : "success";
  return (
    <div className={`statusBar ${status}`}>
      <div>
        {status === "loading" || status === "saving" ? <Spinner size="tiny" /> : <DocumentSave24Regular />}
        <span>{message}</span>
      </div>
      <Badge color={badgeColor}>{badgeLabel}</Badge>
    </div>
  );
}

function StatusBar({ status, dirty, message }: { status: Status; dirty: boolean; message: string }) {
  return (
    <div className={`statusBar ${status}`}>
      <div>
        {status === "loading" || status === "saving" ? <Spinner size="tiny" /> : <DocumentSave24Regular />}
        <span>{message}</span>
      </div>
      <Badge color={dirty ? "warning" : status === "error" ? "danger" : "success"}>{dirty ? "未保存" : status === "error" ? "確認が必要" : "保存済み"}</Badge>
    </div>
  );
}

function headingFor(tab: TabValue) {
  switch (tab) {
    case "companion":
      return "Companion";
    case "jiggle":
      return "Jiggle";
    case "appearance":
      return "Appearance";
    case "updates":
      return "Updates";
    case "advanced":
      return "Advanced";
    default:
      return "Home";
  }
}

function errorMessage(error: unknown) {
  if (error instanceof Error) {
    return error.message;
  }
  return String(error);
}

function normalizeSettings(source: unknown): Settings {
  const merged = { ...defaultSettings, ...(source as Record<string, unknown>) } as Record<string, unknown>;
  const coat = String(merged.CoatColor);
  return {
    PetName: String(merged.PetName || defaultSettings.PetName),
    AssistantEnabled: Boolean(merged.AssistantEnabled),
    LaunchAtLogin: Boolean(merged.LaunchAtLogin),
    QuickKeysEnabled: merged.QuickKeysEnabled !== false,
    QuickKeys: normalizeQuickKeys(merged.QuickKeys),
    QuickScrollEnabled: merged.QuickScrollEnabled !== false,
    QuickScrollLines: clampNumber(numberOrDefault(merged.QuickScrollLines, defaultSettings.QuickScrollLines), 1, 10),
    TargetAssist: false,
    TargetMove: false,
    TargetReturn: false,
    TargetMoveDelayMs: numberOrDefault(merged.TargetMoveDelayMs, defaultSettings.TargetMoveDelayMs),
    TargetRules: [],
    KeepAwakeMode: "off",
    JiggleMode: merged.JiggleMode === "1px" ? "1px" : "off",
    JiggleIdleSec: numberOrDefault(merged.JiggleIdleSec, defaultSettings.JiggleIdleSec),
    SpriteSizePx: numberOrDefault(merged.SpriteSizePx, defaultSettings.SpriteSizePx),
    ShowPetName: Boolean(merged.ShowPetName),
    CoatColor: isCoatColor(coat) ? coat : defaultSettings.CoatColor,
    CompanionOffsetX: numberOrDefault(merged.CompanionOffsetX, defaultSettings.CompanionOffsetX),
    CompanionOffsetY: numberOrDefault(merged.CompanionOffsetY, defaultSettings.CompanionOffsetY),
    EdgeWarpEnabled: Boolean(merged.EdgeWarpEnabled),
    IdleActions: Boolean(merged.IdleActions),
    IdleActionSec: numberOrDefault(merged.IdleActionSec, defaultSettings.IdleActionSec),
    UpdateCheck: Boolean(merged.UpdateCheck),
    UpdateRepo: String(merged.UpdateRepo || defaultSettings.UpdateRepo)
  };
}

function normalizeQuickKeys(source: unknown): QuickKeyID[] {
  if (!Array.isArray(source)) {
    return [...defaultSettings.QuickKeys];
  }
  const seen = new Set<QuickKeyID>();
  const keys: QuickKeyID[] = [];
  for (const item of source) {
    const key = String(item).trim().toLowerCase();
    if (isQuickKeyID(key) && !seen.has(key)) {
      seen.add(key);
      keys.push(key);
    }
  }
  return keys;
}

function signedOffset(value: number) {
  return value > 0 ? `+${value}px` : `${value}px`;
}

function verticalOffsetLabel(value: number) {
  if (value > 0) {
    return `縦位置 +${value}px（上）`;
  }
  if (value < 0) {
    return `縦位置 ${value}px（下）`;
  }
  return "縦位置 0px";
}

function numberOrDefault(value: unknown, fallback: number) {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : fallback;
}

function clampNumber(value: number, min: number, max: number) {
  return Math.min(max, Math.max(min, value));
}

function isCoatColor(value: string): value is CoatColor {
  return Object.prototype.hasOwnProperty.call(coatLabels, value);
}

function isQuickKeyID(value: string): value is QuickKeyID {
  return Object.prototype.hasOwnProperty.call(quickKeyLabels, value);
}

export default App;
