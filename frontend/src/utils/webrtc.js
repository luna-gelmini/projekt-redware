import { get } from 'svelte/store';
import { SendWebRTCSignal } from '../../wailsjs/go/main/App';
import { auth, appContext } from '../stores';

class WebRTCService {
    constructor() {
        this.localStream = null;
        this.peers = new Map(); // peerID -> { connection, stream }
        this.currentChannelId = null;
        this.isMuted = false;
        this.isDeafened = false;
    }

    async startLocalStream() {
        try {
            this.localStream = await navigator.mediaDevices.getUserMedia({ audio: true, video: false });
            this.setMute(this.isMuted); // Apply initial mute state
            return true;
        } catch (err) {
            console.error("Failed to get local stream:", err);
            return false;
        }
    }

    stopLocalStream() {
        if (this.localStream) {
            this.localStream.getTracks().forEach(track => track.stop());
            this.localStream = null;
        }
    }

    async joinChannel(channelId) {
        if (this.currentChannelId === channelId) return;

        // Leave current channel first
        if (this.currentChannelId) {
            this.leaveChannel();
        }

        console.log(`Joining voice channel: ${channelId}`);
        this.currentChannelId = channelId;

        // Get local stream if needed
        if (!this.localStream) {
            const success = await this.startLocalStream();
            if (!success) return;
        }

        // Identify peers in this channel
        // For now, we connect to ALL peers in the server/context. 
        // In a real app, we'd filter by who is in the voice channel.
        // But since we don't have "Voice State" sync yet, we'll just connect to everyone 
        // and let them decide if they want to answer (if they are in the channel).
        // A better approach: Broadcast "I joined channel X" and let others initiate?
        // Or just initiate to everyone.

        const peers = get(appContext).peers;
        const currentUser = get(auth).currentUser;

        peers.forEach(peer => {
            if (peer.user_id !== currentUser.id) {
                this.initiateConnection(peer.user_id);
            }
        });
    }

    leaveChannel() {
        console.log("Leaving voice channel");
        this.currentChannelId = null;
        this.stopLocalStream();

        // Close all connections
        this.peers.forEach(peer => {
            peer.connection.close();
        });
        this.peers.clear();

        // Update UI store? (TODO)
    }

    async initiateConnection(targetId) {
        console.log(`Initiating WebRTC connection to ${targetId}`);
        const pc = this.createPeerConnection(targetId);

        // Add local tracks
        this.localStream.getTracks().forEach(track => {
            pc.addTrack(track, this.localStream);
        });

        try {
            const offer = await pc.createOffer();
            await pc.setLocalDescription(offer);

            await SendWebRTCSignal(targetId, "offer", JSON.stringify(offer), this.currentChannelId);
        } catch (err) {
            console.error(`Error creating offer for ${targetId}:`, err);
        }
    }

    createPeerConnection(targetId) {
        const pc = new RTCPeerConnection({
            iceServers: [
                { urls: 'stun:stun.l.google.com:19302' },
                { urls: 'stun:stun1.l.google.com:19302' }
            ]
        });

        pc.onicecandidate = (event) => {
            if (event.candidate) {
                SendWebRTCSignal(targetId, "candidate", JSON.stringify(event.candidate), this.currentChannelId);
            }
        };

        pc.ontrack = (event) => {
            console.log(`Received remote track from ${targetId}`);
            const remoteStream = event.streams[0];

            // Store peer info
            this.peers.set(targetId, { connection: pc, stream: remoteStream });

            // Notify UI to render audio element (via store or event)
            // For now, let's just create an audio element dynamically
            this.addAudioElement(targetId, remoteStream);
        };

        pc.onconnectionstatechange = () => {
            console.log(`Connection state with ${targetId}: ${pc.connectionState}`);
            if (pc.connectionState === 'disconnected' || pc.connectionState === 'failed') {
                this.removeAudioElement(targetId);
                this.peers.delete(targetId);
            }
        };

        return pc;
    }

    async handleSignal(signal) {
        // signal: { sender_id, type, payload, channel_id }
        if (signal.channel_id !== this.currentChannelId) {
            console.log("Ignored signal for different channel");
            return;
        }

        const senderId = signal.sender_id;
        let peer = this.peers.get(senderId);
        let pc = peer ? peer.connection : null;

        if (!pc) {
            // If it's an offer, we accept and create a new PC
            if (signal.type === "offer") {
                pc = this.createPeerConnection(senderId);
                // Add local tracks if we are in the channel
                if (this.localStream) {
                    this.localStream.getTracks().forEach(track => {
                        pc.addTrack(track, this.localStream);
                    });
                }
                this.peers.set(senderId, { connection: pc, stream: null });
            } else {
                console.warn(`Received ${signal.type} from unknown peer ${senderId}`);
                return;
            }
        }

        try {
            const data = JSON.parse(signal.payload);

            if (signal.type === "offer") {
                await pc.setRemoteDescription(new RTCSessionDescription(data));
                const answer = await pc.createAnswer();
                await pc.setLocalDescription(answer);
                await SendWebRTCSignal(senderId, "answer", JSON.stringify(answer), this.currentChannelId);
            } else if (signal.type === "answer") {
                await pc.setRemoteDescription(new RTCSessionDescription(data));
            } else if (signal.type === "candidate") {
                await pc.addIceCandidate(new RTCIceCandidate(data));
            }
        } catch (err) {
            console.error(`Error handling signal from ${senderId}:`, err);
        }
    }

    addAudioElement(peerId, stream) {
        let audio = document.getElementById(`audio-${peerId}`);
        if (!audio) {
            audio = document.createElement('audio');
            audio.id = `audio-${peerId}`;
            audio.autoplay = true;
            audio.controls = true; // Debugging: Show controls
            audio.style.display = 'block'; // Debugging: Show element
            audio.style.position = 'fixed';
            audio.style.bottom = '10px';
            audio.style.right = '10px';
            audio.style.zIndex = '9999';
            document.body.appendChild(audio);
        }
        audio.srcObject = stream;
    }

    removeAudioElement(peerId) {
        const audio = document.getElementById(`audio-${peerId}`);
        if (audio) {
            audio.srcObject = null;
            audio.remove();
        }
    }

    setMute(muted) {
        this.isMuted = muted;
        if (this.localStream) {
            this.localStream.getAudioTracks().forEach(track => {
                track.enabled = !muted;
            });
        }
    }

    setDeaf(deafened) {
        this.isDeafened = deafened;
        this.peers.forEach(peer => {
            const audio = document.getElementById(`audio-${peer.key}`); // peer.key is undefined here, need to iterate properly
        });

        // Correct iteration
        for (const [id, peer] of this.peers) {
            const audio = document.getElementById(`audio-${id}`);
            if (audio) {
                audio.muted = deafened;
            }
        }
    }
}

export const webRTCService = new WebRTCService();
