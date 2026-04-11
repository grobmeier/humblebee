export namespace guiapp {
	
	export class RunningTimer {
	    workItemName: string;
	    startTimeUTC: number;
	
	    static createFrom(source: any = {}) {
	        return new RunningTimer(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.workItemName = source["workItemName"];
	        this.startTimeUTC = source["startTimeUTC"];
	    }
	}
	export class Dashboard {
	    initialized: boolean;
	    dbPath: string;
	    userEmail: string;
	    running?: RunningTimer;
	    todayTotalSeconds: number;
	
	    static createFrom(source: any = {}) {
	        return new Dashboard(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.initialized = source["initialized"];
	        this.dbPath = source["dbPath"];
	        this.userEmail = source["userEmail"];
	        this.running = this.convertValues(source["running"], RunningTimer);
	        this.todayTotalSeconds = source["todayTotalSeconds"];
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
	
	export class StopResult {
	    workItemName: string;
	    durationSeconds: number;
	    todayTotalSeconds: number;
	
	    static createFrom(source: any = {}) {
	        return new StopResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.workItemName = source["workItemName"];
	        this.durationSeconds = source["durationSeconds"];
	        this.todayTotalSeconds = source["todayTotalSeconds"];
	    }
	}
	export class WorkItem {
	    id: number;
	    name: string;
	    parentId?: number;
	    depth: number;
	
	    static createFrom(source: any = {}) {
	        return new WorkItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.parentId = source["parentId"];
	        this.depth = source["depth"];
	    }
	}

}

