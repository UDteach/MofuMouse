export type KeepAwakeMode = "off" | "os";
export type JiggleMode = "off" | "1px";
export type CoatColor = "agouti" | "gray" | "dark" | "cream" | "white" | "pied";
export type AssistRuleAction = "prefer" | "block";
export type QuickKeyID = "find" | "copy" | "select_all" | "new_tab";

export type AssistRule = {
  Enabled: boolean;
  App: string;
  Window: string;
  Button: string;
  Action: AssistRuleAction;
  Note: string;
};

export type Settings = {
  PetName: string;
  AssistantEnabled: boolean;
  LaunchAtLogin: boolean;
  QuickKeysEnabled: boolean;
  QuickKeys: QuickKeyID[];
  QuickScrollEnabled: boolean;
  QuickScrollLines: number;
  TargetAssist: boolean;
  TargetMove: boolean;
  TargetReturn: boolean;
  TargetMoveDelayMs: number;
  TargetRules: AssistRule[];
  KeepAwakeMode: KeepAwakeMode;
  JiggleMode: JiggleMode;
  JiggleIdleSec: number;
  SpriteSizePx: number;
  ShowPetName: boolean;
  CoatColor: CoatColor;
  CompanionOffsetX: number;
  CompanionOffsetY: number;
  EdgeWarpEnabled: boolean;
  IdleActions: boolean;
  IdleActionSec: number;
  UpdateCheck: boolean;
  UpdateRepo: string;
};

export const defaultSettings: Settings = {
  PetName: "Mofu",
  AssistantEnabled: true,
  LaunchAtLogin: false,
  QuickKeysEnabled: true,
  QuickKeys: ["find", "copy", "select_all", "new_tab"],
  QuickScrollEnabled: true,
  QuickScrollLines: 3,
  TargetAssist: false,
  TargetMove: false,
  TargetReturn: false,
  TargetMoveDelayMs: 900,
  TargetRules: [],
  KeepAwakeMode: "off",
  JiggleMode: "off",
  JiggleIdleSec: 60,
  SpriteSizePx: 32,
  ShowPetName: false,
  CoatColor: "agouti",
  CompanionOffsetX: 0,
  CompanionOffsetY: 0,
  EdgeWarpEnabled: false,
  IdleActions: true,
  IdleActionSec: 45,
  UpdateCheck: false,
  UpdateRepo: "UDteach/MofuMouse"
};

export const coatLabels: Record<CoatColor, string> = {
  agouti: "ノーマル（茶色）",
  gray: "グレー",
  dark: "ダークブラウン",
  cream: "クリーム",
  white: "ホワイト",
  pied: "ぶち模様"
};

export const quickKeyLabels: Record<QuickKeyID, string> = {
  find: "検索",
  copy: "コピー",
  select_all: "全選択",
  new_tab: "新しいタブ"
};

export const quickKeyShortcuts: Record<QuickKeyID, string> = {
  find: "Ctrl+F",
  copy: "Ctrl+C",
  select_all: "Ctrl+A",
  new_tab: "Ctrl+T"
};
