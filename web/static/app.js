class WebRTCClient {
    constructor() {
        this.ws = null;
        this.localStream = null;
        this.peerConnections = new Map();
        this.iceServers = [];
        this.currentRoom = null;
        this.userId = null;
        
        this.initializeElements();
        this.setupEventListeners();
        this.connectWebSocket();
    }

    initializeElements() {
        this.elements = {
            roomId: document.getElementById('roomId'),
            joinBtn: document.getElementById('joinBtn'),
            leaveBtn: document.getElementById('leaveBtn'),
            startVideoBtn: document.getElementById('startVideoBtn'),
            stopVideoBtn: document.getElementById('stopVideoBtn'),
            clearLogsBtn: document.getElementById('clearLogsBtn'),
            connectionStatus: document.getElementById('connectionStatus'),
            roomStatus: document.getElementById('roomStatus'),
            localVideo: document.getElementById('localVideo'),
            remoteVideos: document.getElementById('remoteVideos'),
            logContainer: document.getElementById('logContainer')
        };
    }

    setupEventListeners() {
        this.elements.joinBtn.addEventListener('click', () => this.joinRoom());
        this.elements.leaveBtn.addEventListener('click', () => this.leaveRoom());
        this.elements.startVideoBtn.addEventListener('click', () => this.startVideo());
        this.elements.stopVideoBtn.addEventListener('click', () => this.stopVideo());
        this.elements.clearLogsBtn.addEventListener('click', () => this.clearLogs());
        
        // Enter key support for room input
        this.elements.roomId.addEventListener('keypress', (e) => {
            if (e.key === 'Enter' && !this.elements.joinBtn.disabled) {
                this.joinRoom();
            }
        });
    }

    connectWebSocket() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws`;
        
        this.log('Connecting to WebSocket...', 'info');
        
        this.ws = new WebSocket(wsUrl);
        
        this.ws.onopen = () => {
            this.log('WebSocket connected', 'success');
            this.updateConnectionStatus(true);
        };
        
        this.ws.onmessage = (event) => {
            this.handleWebSocketMessage(JSON.parse(event.data));
        };
        
        this.ws.onclose = () => {
            this.log('WebSocket disconnected', 'warning');
            this.updateConnectionStatus(false);
            this.cleanup();
            
            // Attempt to reconnect after 3 seconds
            setTimeout(() => {
                if (!this.ws || this.ws.readyState === WebSocket.CLOSED) {
                    this.connectWebSocket();
                }
            }, 3000);
        };
        
        this.ws.onerror = (error) => {
            this.log(`WebSocket error: ${error}`, 'error');
        };
    }

    async handleWebSocketMessage(message) {
        this.log(`Received message: ${JSON.stringify(message)}`, 'info');
        
        switch (message.type) {
            case 'stun_config':
                this.iceServers = message.data.iceServers;
                this.log(`STUN/TURN servers configured: ${JSON.stringify(this.iceServers)}`, 'info');
                break;
                
            case 'user_joined':
                this.log(`Processing user_joined message for user: ${message.user_id}`, 'info');
                await this.handleUserJoined(message);
                break;
                
            case 'user_left':
                this.log(`Processing user_left message for user: ${message.user_id}`, 'info');
                this.handleUserLeft(message);
                break;
                
            case 'offer':
                this.log(`Processing offer from user: ${message.user_id}`, 'info');
                await this.handleOffer(message);
                break;
                
            case 'answer':
                this.log(`Processing answer from user: ${message.user_id}`, 'info');
                await this.handleAnswer(message);
                break;
                
            case 'ice_candidate':
                this.log(`Processing ICE candidate from user: ${message.user_id}`, 'info');
                await this.handleIceCandidate(message);
                break;
                
            case 'room_full':
                this.log('Room is full', 'error');
                alert('Room is full. Please try another room.');
                break;
                
            case 'error':
                try {
                    const errorData = typeof message.data === 'string' ? JSON.parse(message.data) : message.data;
                    this.log(`Server error: ${errorData.message || 'Unknown error'}`, 'error');
                } catch (e) {
                    this.log(`Server error: ${JSON.stringify(message.data)}`, 'error');
                }
                break;
                
            default:
                this.log(`Unknown message type: ${message.type}, full message: ${JSON.stringify(message)}`, 'warning');
        }
    }

    async joinRoom() {
        const roomId = this.elements.roomId.value.trim();
        if (!roomId) {
            alert('Please enter a room ID');
            return;
        }
        
        if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
            alert('WebSocket not connected');
            return;
        }
        
        const message = {
            type: 'join_room',
            data: { room_id: roomId }
        };
        
        this.ws.send(JSON.stringify(message));
        this.currentRoom = roomId;
        this.updateRoomStatus(`Joined: ${roomId}`);
        
        this.elements.joinBtn.disabled = true;
        this.elements.leaveBtn.disabled = false;
        this.elements.roomId.disabled = true;
        
        this.log(`Joining room: ${roomId}`, 'info');
    }

    leaveRoom() {
        if (!this.currentRoom) return;
        
        const message = {
            type: 'leave_room',
            room_id: this.currentRoom
        };
        
        this.ws.send(JSON.stringify(message));
        this.cleanup();
        
        this.log(`Left room: ${this.currentRoom}`, 'info');
    }

    async startVideo() {
        try {
            this.localStream = await navigator.mediaDevices.getUserMedia({
                video: true,
                audio: true
            });
            
            this.elements.localVideo.srcObject = this.localStream;
            this.elements.startVideoBtn.disabled = true;
            this.elements.stopVideoBtn.disabled = false;
            
            this.log('Local video started', 'success');
            
            // Add tracks to existing peer connections
            for (const [userId, pc] of this.peerConnections) {
                this.log(`Adding tracks to existing peer connection with ${userId}`, 'info');
                this.localStream.getTracks().forEach(track => {
                    try {
                        pc.addTrack(track, this.localStream);
                        this.log(`Added ${track.kind} track to peer connection with ${userId}`, 'info');
                    } catch (error) {
                        this.log(`Failed to add ${track.kind} track to peer connection with ${userId}: ${error.message}`, 'warning');
                    }
                });
            }
            
        } catch (error) {
            this.log(`Failed to start video: ${error.message}`, 'error');
            alert('Failed to access camera/microphone');
            throw error;
        }
    }

    stopVideo() {
        if (this.localStream) {
            this.localStream.getTracks().forEach(track => track.stop());
            this.localStream = null;
            this.elements.localVideo.srcObject = null;
        }
        
        this.elements.startVideoBtn.disabled = false;
        this.elements.stopVideoBtn.disabled = true;
        
        this.log('Local video stopped', 'info');
    }

    async handleUserJoined(message) {
        const data = typeof message.data === 'string' ? JSON.parse(message.data) : message.data;
        this.log(`User joined: ${message.user_id}`, 'success');
        
        // Set our own user ID if this is our join confirmation
        if (!this.userId) {
            this.userId = message.user_id;
            this.log(`My user ID: ${this.userId}`, 'info');
            
            // Auto-start video when we join
            if (!this.localStream) {
                await this.startVideo();
            }
            
            // If there are existing users in the room, create connections to them
            if (data.users && data.users.length > 1) {
                for (const userId of data.users) {
                    if (userId !== this.userId && !this.peerConnections.has(userId)) {
                        this.log(`Creating peer connection to existing user ${userId}`, 'info');
                        // Wait a bit to ensure local stream is ready
                        setTimeout(async () => {
                            await this.createPeerConnection(userId, true); // We initiate the connection
                        }, 1500);
                    }
                }
            }
            
            return; // Don't create peer connection to ourselves
        }
        
        // Create peer connection for other users joining after us
        if (message.user_id !== this.userId && !this.peerConnections.has(message.user_id)) {
            this.log(`Creating peer connection to new user ${message.user_id}`, 'info');
            // Wait a bit to ensure both users have their local streams ready
            setTimeout(async () => {
                await this.createPeerConnection(message.user_id, false); // They will initiate
            }, 1000);
        }
    }

    handleUserLeft(message) {
        this.log(`User left: ${message.user_id}`, 'warning');
        this.removePeerConnection(message.user_id);
    }

    async createPeerConnection(userId, isInitiator = false) {
        this.log(`Creating peer connection to ${userId}, isInitiator: ${isInitiator}`, 'info');
        
        const pc = new RTCPeerConnection({ iceServers: this.iceServers });
        this.peerConnections.set(userId, pc);
        
        // Handle remote stream
        pc.ontrack = (event) => {
            this.log(`Received remote stream from ${userId}`, 'success');
            this.log(`Remote stream has ${event.streams[0].getTracks().length} tracks`, 'info');
            this.addRemoteVideo(userId, event.streams[0]);
        };
        
        // Handle ICE candidates
        pc.onicecandidate = (event) => {
            if (event.candidate) {
                this.log(`Sending ICE candidate to ${userId}`, 'info');
                this.sendMessage({
                    type: 'ice_candidate',
                    target_id: userId,
                    data: JSON.stringify({
                        candidate: event.candidate.candidate,
                        sdpMid: event.candidate.sdpMid,
                        sdpMLineIndex: event.candidate.sdpMLineIndex
                    })
                });
            } else {
                this.log(`ICE gathering complete for ${userId}`, 'info');
            }
        };
        
        // Handle connection state changes
        pc.onconnectionstatechange = () => {
            this.log(`Connection state with ${userId}: ${pc.connectionState}`, 'info');
            if (pc.connectionState === 'failed') {
                this.log(`Connection failed with ${userId}, attempting to restart ICE`, 'warning');
                pc.restartIce();
            }
        };
        
        // Handle ICE connection state changes
        pc.oniceconnectionstatechange = () => {
            this.log(`ICE connection state with ${userId}: ${pc.iceConnectionState}`, 'info');
        };
        
        // Ensure we have local stream before adding tracks
        if (!this.localStream) {
            this.log(`No local stream available, starting video for peer connection with ${userId}`, 'info');
            await this.startVideo();
        }
        
        // Add local stream tracks
        if (this.localStream) {
            this.localStream.getTracks().forEach(track => {
                this.log(`Adding local track ${track.kind} to peer connection with ${userId}`, 'info');
                pc.addTrack(track, this.localStream);
            });
        } else {
            this.log(`Warning: Still no local stream available for peer connection with ${userId}`, 'warning');
        }
        
        if (isInitiator) {
            // Wait a bit to ensure everything is set up
            setTimeout(async () => {
                try {
                    this.log(`Creating offer for ${userId}`, 'info');
                    const offer = await pc.createOffer();
                    await pc.setLocalDescription(offer);
                    
                    this.log(`Sending offer to ${userId}`, 'info');
                    this.sendMessage({
                        type: 'offer',
                        target_id: userId,
                        data: JSON.stringify({
                            sdp: offer.sdp,
                            type: offer.type
                        })
                    });
                } catch (error) {
                    this.log(`Failed to create/send offer to ${userId}: ${error.message}`, 'error');
                }
            }, 1000);
        }
    }

    async handleOffer(message) {
        const offerData = JSON.parse(message.data);
        let pc = this.peerConnections.get(message.user_id);
        
        if (!pc) {
            await this.createPeerConnection(message.user_id, false);
            pc = this.peerConnections.get(message.user_id);
        }
        
        await pc.setRemoteDescription(new RTCSessionDescription(offerData));
        
        const answer = await pc.createAnswer();
        await pc.setLocalDescription(answer);
        
        this.sendMessage({
            type: 'answer',
            target_id: message.user_id,
            data: JSON.stringify({
                sdp: answer.sdp,
                type: answer.type
            })
        });
    }

    async handleAnswer(message) {
        const answerData = JSON.parse(message.data);
        const pc = this.peerConnections.get(message.user_id);
        
        if (pc) {
            await pc.setRemoteDescription(new RTCSessionDescription(answerData));
        }
    }

    async handleIceCandidate(message) {
        const candidateData = JSON.parse(message.data);
        const pc = this.peerConnections.get(message.user_id);
        
        if (pc) {
            await pc.addIceCandidate(new RTCIceCandidate(candidateData));
        }
    }

    addRemoteVideo(userId, stream) {
        this.log(`Adding remote video for user ${userId}`, 'info');
        
        // Remove existing video if any
        const existingContainer = document.getElementById(`container-${userId}`);
        if (existingContainer) {
            existingContainer.remove();
            this.log(`Removed existing video container for user ${userId}`, 'info');
        }
        
        const videoContainer = document.createElement('div');
        videoContainer.className = 'remote-video-container';
        videoContainer.id = `container-${userId}`;
        
        const video = document.createElement('video');
        video.id = `video-${userId}`;
        video.autoplay = true;
        video.playsinline = true;
        video.muted = false; // Allow audio from remote users
        video.srcObject = stream;
        
        // Add event listeners for video debugging
        video.addEventListener('loadedmetadata', () => {
            this.log(`Remote video metadata loaded for ${userId}`, 'success');
        });
        
        video.addEventListener('canplay', () => {
            this.log(`Remote video can play for ${userId}`, 'success');
        });
        
        video.addEventListener('error', (e) => {
            this.log(`Remote video error for ${userId}: ${e.message}`, 'error');
        });
        
        const label = document.createElement('div');
        label.className = 'remote-video-label';
        label.textContent = userId.substring(0, 8);
        
        videoContainer.appendChild(video);
        videoContainer.appendChild(label);
        this.elements.remoteVideos.appendChild(videoContainer);
        
        this.log(`Remote video element created and added for user ${userId}`, 'success');
        
        // Log stream information
        if (stream) {
            const tracks = stream.getTracks();
            this.log(`Remote stream has ${tracks.length} tracks: ${tracks.map(t => t.kind).join(', ')}`, 'info');
        }
    }

    removePeerConnection(userId) {
        const pc = this.peerConnections.get(userId);
        if (pc) {
            pc.close();
            this.peerConnections.delete(userId);
        }
        
        const videoContainer = document.getElementById(`container-${userId}`);
        if (videoContainer) {
            videoContainer.remove();
        }
    }

    sendMessage(message) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify(message));
        }
    }

    cleanup() {
        this.log('Cleaning up client state', 'info');
        
        // Close all peer connections
        for (const [userId, pc] of this.peerConnections) {
            this.log(`Closing peer connection with ${userId}`, 'info');
            pc.close();
        }
        this.peerConnections.clear();
        
        // Remove all remote videos
        this.elements.remoteVideos.innerHTML = '';
        
        // Reset client state completely
        this.currentRoom = null;
        this.userId = null; // Reset user ID so we get a fresh one on rejoin
        
        // Reset UI state
        this.elements.joinBtn.disabled = false;
        this.elements.leaveBtn.disabled = true;
        this.elements.roomId.disabled = false;
        this.updateRoomStatus('');
        
        this.log('Client state cleaned up successfully', 'info');
    }

    updateConnectionStatus(connected) {
        this.elements.connectionStatus.textContent = connected ? 'Connected' : 'Disconnected';
        this.elements.connectionStatus.className = connected ? 'connected' : '';
    }

    updateRoomStatus(status) {
        this.elements.roomStatus.textContent = status;
    }

    log(message, type = 'info') {
        const timestamp = new Date().toLocaleTimeString();
        const logEntry = document.createElement('div');
        logEntry.className = `log-entry ${type}`;
        
        const timestampSpan = document.createElement('span');
        timestampSpan.className = 'log-timestamp';
        timestampSpan.textContent = `[${timestamp}] `;
        
        logEntry.appendChild(timestampSpan);
        logEntry.appendChild(document.createTextNode(message));
        
        this.elements.logContainer.appendChild(logEntry);
        this.elements.logContainer.scrollTop = this.elements.logContainer.scrollHeight;
        
        console.log(`[${timestamp}] ${message}`);
    }

    clearLogs() {
        this.elements.logContainer.innerHTML = '';
    }
}

// Initialize the WebRTC client when the page loads
document.addEventListener('DOMContentLoaded', () => {
    new WebRTCClient();
});
