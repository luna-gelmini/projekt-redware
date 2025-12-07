export namespace p2p {
	
	export class Peer {
	    UserID: string;
	    Username: string;
	
	    static createFrom(source: any = {}) {
	        return new Peer(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.UserID = source["UserID"];
	        this.Username = source["Username"];
	    }
	}
	export class PeerInfo {
	    user_id: string;
	    username: string;
	    ip: string;
	    port: number;
	    avatar: string;
	    status: string;
	    status_message: string;
	    server_name: string;
	    server_icon: string;
	    server_description: string;
	
	    static createFrom(source: any = {}) {
	        return new PeerInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.user_id = source["user_id"];
	        this.username = source["username"];
	        this.ip = source["ip"];
	        this.port = source["port"];
	        this.avatar = source["avatar"];
	        this.status = source["status"];
	        this.status_message = source["status_message"];
	        this.server_name = source["server_name"];
	        this.server_icon = source["server_icon"];
	        this.server_description = source["server_description"];
	    }
	}

}

export namespace proto {
	
	export class Attachment {
	    filename?: string;
	    url?: string;
	    mime_type?: string;
	    size?: number;
	
	    static createFrom(source: any = {}) {
	        return new Attachment(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.filename = source["filename"];
	        this.url = source["url"];
	        this.mime_type = source["mime_type"];
	        this.size = source["size"];
	    }
	}
	export class ChannelMetadata {
	    channel_id?: string;
	    owner_id?: string;
	    topic?: string;
	    is_private?: boolean;
	    software_agent?: string;
	    is_modded?: boolean;
	    signature?: string;
	    icon?: string;
	    description?: string;
	    type?: string;
	    name?: string;
	
	    static createFrom(source: any = {}) {
	        return new ChannelMetadata(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.channel_id = source["channel_id"];
	        this.owner_id = source["owner_id"];
	        this.topic = source["topic"];
	        this.is_private = source["is_private"];
	        this.software_agent = source["software_agent"];
	        this.is_modded = source["is_modded"];
	        this.signature = source["signature"];
	        this.icon = source["icon"];
	        this.description = source["description"];
	        this.type = source["type"];
	        this.name = source["name"];
	    }
	}
	export class ChatMessage {
	    id?: string;
	    content?: string;
	    timestamp?: number;
	    reply_to_id?: string;
	    attachments?: Attachment[];
	    channel_id?: string;
	    sender_id?: string;
	    sender_username?: string;
	
	    static createFrom(source: any = {}) {
	        return new ChatMessage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.content = source["content"];
	        this.timestamp = source["timestamp"];
	        this.reply_to_id = source["reply_to_id"];
	        this.attachments = this.convertValues(source["attachments"], Attachment);
	        this.channel_id = source["channel_id"];
	        this.sender_id = source["sender_id"];
	        this.sender_username = source["sender_username"];
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
	export class Member {
	    user_id?: string;
	    username?: string;
	    avatar?: string;
	    status?: string;
	
	    static createFrom(source: any = {}) {
	        return new Member(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.user_id = source["user_id"];
	        this.username = source["username"];
	        this.avatar = source["avatar"];
	        this.status = source["status"];
	    }
	}

}

export namespace store {
	
	export class JoinedNode {
	    node_id: string;
	    address: string;
	    name: string;
	    description: string;
	    icon: string;
	    owner_id: string;
	    // Go type: time
	    joined_at: any;
	    is_server: boolean;
	
	    static createFrom(source: any = {}) {
	        return new JoinedNode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.node_id = source["node_id"];
	        this.address = source["address"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.icon = source["icon"];
	        this.owner_id = source["owner_id"];
	        this.joined_at = this.convertValues(source["joined_at"], null);
	        this.is_server = source["is_server"];
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
	export class Message {
	    id: number;
	    sender_id: string;
	    sender_username: string;
	    content: string;
	    channel_id: string;
	    // Go type: time
	    timestamp: any;
	    is_outgoing: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Message(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.sender_id = source["sender_id"];
	        this.sender_username = source["sender_username"];
	        this.content = source["content"];
	        this.channel_id = source["channel_id"];
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.is_outgoing = source["is_outgoing"];
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

