declare module "../wailsjs/go/main/ControlCenterApp" {
  import type { Settings } from "./settings";

  export function ConfigPath(): Promise<string>;
  export function GetSettings(): Promise<Settings>;
  export function OpenConfigFile(): Promise<void>;
  export function ResetSettings(): Promise<Settings>;
  export function SaveSettings(settings: Settings): Promise<Settings>;
}
