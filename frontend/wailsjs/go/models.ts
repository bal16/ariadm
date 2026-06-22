export namespace config {
	
	export class AppConfig {
	    default_download_path: string;
	    speed_limit: number;
	    max_concurrent_tasks: number;
	    minimize_to_tray: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AppConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.default_download_path = source["default_download_path"];
	        this.speed_limit = source["speed_limit"];
	        this.max_concurrent_tasks = source["max_concurrent_tasks"];
	        this.minimize_to_tray = source["minimize_to_tray"];
	    }
	}

}

export namespace task {
	
	export class Task {
	    id: string;
	    gid: string;
	    url: string;
	    file_name: string;
	    total_length: number;
	    completed_length: number;
	    speed: number;
	    status: string;
	    // Go type: time
	    created_at: any;
	
	    static createFrom(source: any = {}) {
	        return new Task(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.gid = source["gid"];
	        this.url = source["url"];
	        this.file_name = source["file_name"];
	        this.total_length = source["total_length"];
	        this.completed_length = source["completed_length"];
	        this.speed = source["speed"];
	        this.status = source["status"];
	        this.created_at = this.convertValues(source["created_at"], null);
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

