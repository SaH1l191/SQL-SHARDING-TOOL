export namespace main {
	
	export class WailsConfig {
	    WindowTitle: string;
	    WindowWidth: number;
	    WindowHeight: number;
	    DevMode: boolean;
	
	    static createFrom(source: any = {}) {
	        return new WailsConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.WindowTitle = source["WindowTitle"];
	        this.WindowWidth = source["WindowWidth"];
	        this.WindowHeight = source["WindowHeight"];
	        this.DevMode = source["DevMode"];
	    }
	}

}

export namespace repository {
	
	export class Project {
	    id: string;
	    name: string;
	    description: string;
	    status: string;
	    shard_count: number;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	
	    static createFrom(source: any = {}) {
	        return new Project(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.status = source["status"];
	        this.shard_count = source["shard_count"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
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
	export class ProjectSchema {
	    id: string;
	    project_id: string;
	    version: number;
	    state: string;
	    ddl_sql: string;
	    error_message?: string;
	    created_at: string;
	    committed_at?: string;
	    applied_at?: string;
	
	    static createFrom(source: any = {}) {
	        return new ProjectSchema(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.project_id = source["project_id"];
	        this.version = source["version"];
	        this.state = source["state"];
	        this.ddl_sql = source["ddl_sql"];
	        this.error_message = source["error_message"];
	        this.created_at = source["created_at"];
	        this.committed_at = source["committed_at"];
	        this.applied_at = source["applied_at"];
	    }
	}
	export class SchemaExecutionStatus {
	    id: string;
	    schema_id: string;
	    shard_id: string;
	    state: string;
	    err_msg: string;
	    // Go type: time
	    executed_at: any;
	
	    static createFrom(source: any = {}) {
	        return new SchemaExecutionStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.schema_id = source["schema_id"];
	        this.shard_id = source["shard_id"];
	        this.state = source["state"];
	        this.err_msg = source["err_msg"];
	        this.executed_at = this.convertValues(source["executed_at"], null);
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
	export class Shard {
	    id: string;
	    project_id: string;
	    shard_index: number;
	    status: string;
	    // Go type: time
	    created_at: any;
	
	    static createFrom(source: any = {}) {
	        return new Shard(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.project_id = source["project_id"];
	        this.shard_index = source["shard_index"];
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
	export class ShardConnection {
	    shard_id: string;
	    host: string;
	    port: number;
	    database_name: string;
	    username: string;
	    password: string;
	    created_at: string;
	    updated_at: string;
	
	    static createFrom(source: any = {}) {
	        return new ShardConnection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.shard_id = source["shard_id"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.database_name = source["database_name"];
	        this.username = source["username"];
	        this.password = source["password"];
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	    }
	}
	export class ShardKeyRecord {
	    TableName: string;
	    ShardKeyColumn: string;
	    IsManual: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ShardKeyRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.TableName = source["TableName"];
	        this.ShardKeyColumn = source["ShardKeyColumn"];
	        this.IsManual = source["IsManual"];
	    }
	}
	export class ShardKeys {
	    project_id: string;
	    table_name: string;
	    shard_key_column: string;
	    is_manual_override: boolean;
	    // Go type: time
	    updated_at: any;
	
	    static createFrom(source: any = {}) {
	        return new ShardKeys(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.project_id = source["project_id"];
	        this.table_name = source["table_name"];
	        this.shard_key_column = source["shard_key_column"];
	        this.is_manual_override = source["is_manual_override"];
	        this.updated_at = this.convertValues(source["updated_at"], null);
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

