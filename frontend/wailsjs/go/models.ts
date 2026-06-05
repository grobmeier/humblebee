export namespace guiapp {
	
	export class CreateTimeEntryRequest {
	    id: number;
	    workItemId: number;
	    description: string;
	    startDate: string;
	    startTime: string;
	    endDate: string;
	    endTime: string;
	    untilMidnight: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CreateTimeEntryRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.workItemId = source["workItemId"];
	        this.description = source["description"];
	        this.startDate = source["startDate"];
	        this.startTime = source["startTime"];
	        this.endDate = source["endDate"];
	        this.endTime = source["endTime"];
	        this.untilMidnight = source["untilMidnight"];
	    }
	}
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
	export class ReportRequest {
	    mode: string;
	    month: number;
	    year: number;
	    startDate: string;
	    endDate: string;
	    projectId: number;
	
	    static createFrom(source: any = {}) {
	        return new ReportRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mode = source["mode"];
	        this.month = source["month"];
	        this.year = source["year"];
	        this.startDate = source["startDate"];
	        this.endDate = source["endDate"];
	        this.projectId = source["projectId"];
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
	export class Stopwatch {
	    id: number;
	    workItemId?: number;
	    workItemName: string;
	    startDate: string;
	    startTime: string;
	    endDate: string;
	    endTime: string;
	    durationSeconds: number;
	    running: boolean;
	    conflicting: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Stopwatch(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.workItemId = source["workItemId"];
	        this.workItemName = source["workItemName"];
	        this.startDate = source["startDate"];
	        this.startTime = source["startTime"];
	        this.endDate = source["endDate"];
	        this.endTime = source["endTime"];
	        this.durationSeconds = source["durationSeconds"];
	        this.running = source["running"];
	        this.conflicting = source["conflicting"];
	    }
	}
	export class TimeEntry {
	    id: number;
	    workItemId?: number;
	    description: string;
	    startDate: string;
	    startTime: string;
	    endDate: string;
	    endTime: string;
	    durationSeconds: number;
	
	    static createFrom(source: any = {}) {
	        return new TimeEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.workItemId = source["workItemId"];
	        this.description = source["description"];
	        this.startDate = source["startDate"];
	        this.startTime = source["startTime"];
	        this.endDate = source["endDate"];
	        this.endTime = source["endTime"];
	        this.durationSeconds = source["durationSeconds"];
	    }
	}
	export class TimeDay {
	    date: string;
	    entries: TimeEntry[];
	    totalSeconds: number;
	    projectSeconds: number;
	    absenceSeconds: number;
	    workSeconds: number;
	    breakSeconds: number;
	
	    static createFrom(source: any = {}) {
	        return new TimeDay(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.date = source["date"];
	        this.entries = this.convertValues(source["entries"], TimeEntry);
	        this.totalSeconds = source["totalSeconds"];
	        this.projectSeconds = source["projectSeconds"];
	        this.absenceSeconds = source["absenceSeconds"];
	        this.workSeconds = source["workSeconds"];
	        this.breakSeconds = source["breakSeconds"];
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
	
	export class TimesheetDailyRow {
	    date: string;
	    totalSeconds: number;
	    totalDuration: string;
	    projectSeconds: number;
	    projectDuration: string;
	
	    static createFrom(source: any = {}) {
	        return new TimesheetDailyRow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.date = source["date"];
	        this.totalSeconds = source["totalSeconds"];
	        this.totalDuration = source["totalDuration"];
	        this.projectSeconds = source["projectSeconds"];
	        this.projectDuration = source["projectDuration"];
	    }
	}
	export class TimesheetProjectRow {
	    projectId: number;
	    projectName: string;
	    durationSeconds: number;
	    duration: string;
	
	    static createFrom(source: any = {}) {
	        return new TimesheetProjectRow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.projectName = source["projectName"];
	        this.durationSeconds = source["durationSeconds"];
	        this.duration = source["duration"];
	    }
	}
	export class TimesheetReport {
	    empty: boolean;
	    userName: string;
	    projectRows: TimesheetProjectRow[];
	    dailyRows: TimesheetDailyRow[];
	    totalSeconds: number;
	    totalDuration: string;
	
	    static createFrom(source: any = {}) {
	        return new TimesheetReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.empty = source["empty"];
	        this.userName = source["userName"];
	        this.projectRows = this.convertValues(source["projectRows"], TimesheetProjectRow);
	        this.dailyRows = this.convertValues(source["dailyRows"], TimesheetDailyRow);
	        this.totalSeconds = source["totalSeconds"];
	        this.totalDuration = source["totalDuration"];
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
	export class WorkItem {
	    id: number;
	    name: string;
	    parentId?: number;
	    depth: number;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new WorkItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.parentId = source["parentId"];
	        this.depth = source["depth"];
	        this.status = source["status"];
	    }
	}
	export class WorktimeReportRow {
	    projectId: number;
	    projectName: string;
	    taskId: number;
	    taskName: string;
	    description: string;
	    date: string;
	    startTime: string;
	    endTime: string;
	    durationSeconds: number;
	    duration: string;
	
	    static createFrom(source: any = {}) {
	        return new WorktimeReportRow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.projectName = source["projectName"];
	        this.taskId = source["taskId"];
	        this.taskName = source["taskName"];
	        this.description = source["description"];
	        this.date = source["date"];
	        this.startTime = source["startTime"];
	        this.endTime = source["endTime"];
	        this.durationSeconds = source["durationSeconds"];
	        this.duration = source["duration"];
	    }
	}
	export class WorktimeByMonthReport {
	    empty: boolean;
	    rows: WorktimeReportRow[];
	    totalSeconds: number;
	    totalDuration: string;
	
	    static createFrom(source: any = {}) {
	        return new WorktimeByMonthReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.empty = source["empty"];
	        this.rows = this.convertValues(source["rows"], WorktimeReportRow);
	        this.totalSeconds = source["totalSeconds"];
	        this.totalDuration = source["totalDuration"];
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
	export class WorktimeProjectGroup {
	    projectId: number;
	    projectName: string;
	    rows: WorktimeReportRow[];
	    totalSeconds: number;
	    totalDuration: string;
	
	    static createFrom(source: any = {}) {
	        return new WorktimeProjectGroup(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.projectName = source["projectName"];
	        this.rows = this.convertValues(source["rows"], WorktimeReportRow);
	        this.totalSeconds = source["totalSeconds"];
	        this.totalDuration = source["totalDuration"];
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
	export class WorktimeGroupedByProjectReport {
	    empty: boolean;
	    groups: WorktimeProjectGroup[];
	    totalSeconds: number;
	    totalDuration: string;
	
	    static createFrom(source: any = {}) {
	        return new WorktimeGroupedByProjectReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.empty = source["empty"];
	        this.groups = this.convertValues(source["groups"], WorktimeProjectGroup);
	        this.totalSeconds = source["totalSeconds"];
	        this.totalDuration = source["totalDuration"];
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
	
	
	export class WorktimeTaskDetailRow {
	    projectId: number;
	    projectName: string;
	    taskId: number;
	    taskName: string;
	    durationSeconds: number;
	    duration: string;
	
	    static createFrom(source: any = {}) {
	        return new WorktimeTaskDetailRow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.projectName = source["projectName"];
	        this.taskId = source["taskId"];
	        this.taskName = source["taskName"];
	        this.durationSeconds = source["durationSeconds"];
	        this.duration = source["duration"];
	    }
	}
	export class WorktimeTaskDetailsReport {
	    empty: boolean;
	    rows: WorktimeTaskDetailRow[];
	    totalSeconds: number;
	    totalDuration: string;
	
	    static createFrom(source: any = {}) {
	        return new WorktimeTaskDetailsReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.empty = source["empty"];
	        this.rows = this.convertValues(source["rows"], WorktimeTaskDetailRow);
	        this.totalSeconds = source["totalSeconds"];
	        this.totalDuration = source["totalDuration"];
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

