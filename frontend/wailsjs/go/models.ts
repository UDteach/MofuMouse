export namespace core {
	
	export class AssistRule {
	    Enabled: boolean;
	    App: string;
	    Window: string;
	    Button: string;
	    Action: string;
	    Note: string;
	
	    static createFrom(source: any = {}) {
	        return new AssistRule(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Enabled = source["Enabled"];
	        this.App = source["App"];
	        this.Window = source["Window"];
	        this.Button = source["Button"];
	        this.Action = source["Action"];
	        this.Note = source["Note"];
	    }
	}
	export class Settings {
	    PetName: string;
	    AssistantEnabled: boolean;
	    LaunchAtLogin: boolean;
	    QuickKeysEnabled: boolean;
	    QuickKeys: string[];
	    QuickScrollEnabled: boolean;
	    QuickScrollLines: number;
	    TargetAssist: boolean;
	    TargetMove: boolean;
	    TargetReturn: boolean;
	    TargetMoveDelayMs: number;
	    TargetRules: AssistRule[];
	    KeepAwakeMode: string;
	    JiggleMode: string;
	    JiggleIdleSec: number;
	    SpriteSizePx: number;
	    ShowPetName: boolean;
	    CoatColor: string;
	    CompanionOffsetX: number;
	    CompanionOffsetY: number;
	    EdgeWarpEnabled: boolean;
	    IdleActions: boolean;
	    IdleActionSec: number;
	    UpdateCheck: boolean;
	    UpdateRepo: string;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.PetName = source["PetName"];
	        this.AssistantEnabled = source["AssistantEnabled"];
	        this.LaunchAtLogin = source["LaunchAtLogin"];
	        this.QuickKeysEnabled = source["QuickKeysEnabled"];
	        this.QuickKeys = source["QuickKeys"];
	        this.QuickScrollEnabled = source["QuickScrollEnabled"];
	        this.QuickScrollLines = source["QuickScrollLines"];
	        this.TargetAssist = source["TargetAssist"];
	        this.TargetMove = source["TargetMove"];
	        this.TargetReturn = source["TargetReturn"];
	        this.TargetMoveDelayMs = source["TargetMoveDelayMs"];
	        this.TargetRules = this.convertValues(source["TargetRules"], AssistRule);
	        this.KeepAwakeMode = source["KeepAwakeMode"];
	        this.JiggleMode = source["JiggleMode"];
	        this.JiggleIdleSec = source["JiggleIdleSec"];
	        this.SpriteSizePx = source["SpriteSizePx"];
	        this.ShowPetName = source["ShowPetName"];
	        this.CoatColor = source["CoatColor"];
	        this.CompanionOffsetX = source["CompanionOffsetX"];
	        this.CompanionOffsetY = source["CompanionOffsetY"];
	        this.EdgeWarpEnabled = source["EdgeWarpEnabled"];
	        this.IdleActions = source["IdleActions"];
	        this.IdleActionSec = source["IdleActionSec"];
	        this.UpdateCheck = source["UpdateCheck"];
	        this.UpdateRepo = source["UpdateRepo"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

