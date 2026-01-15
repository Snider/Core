import { Component, OnInit, OnDestroy, ElementRef, ViewChild } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';

interface Message {
  role: 'user' | 'assistant' | 'system';
  content: string;
  timestamp: Date;
}

interface WsMessage {
  type: string;
  channel?: string;
  processId?: string;
  data?: any;
  timestamp: string;
}

@Component({
  selector: 'claude-panel',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="claude-panel">
      <div class="panel-header">
        <span class="panel-title">
          <i class="fa-regular fa-message-bot"></i>
          CLAUDE
        </span>
        <span class="connection-status" [class.connected]="connected">
          {{ connected ? 'Connected' : 'Disconnected' }}
        </span>
        <button class="panel-btn" (click)="clearMessages()" title="Clear">
          <i class="fa-regular fa-trash"></i>
        </button>
      </div>

      <div class="messages-container" #messagesContainer>
        @if (messages.length === 0) {
          <div class="empty-state">
            <i class="fa-regular fa-comments"></i>
            <p>No messages yet</p>
            <p class="hint">Start a conversation with Claude</p>
          </div>
        }
        @for (msg of messages; track msg.timestamp) {
          <div class="message" [class]="msg.role">
            <div class="message-header">
              <span class="role-icon">
                @if (msg.role === 'user') {
                  <i class="fa-regular fa-user"></i>
                } @else if (msg.role === 'assistant') {
                  <i class="fa-regular fa-robot"></i>
                } @else {
                  <i class="fa-regular fa-gear"></i>
                }
              </span>
              <span class="role-label">{{ msg.role === 'assistant' ? 'Claude' : msg.role }}</span>
              <span class="timestamp">{{ formatTime(msg.timestamp) }}</span>
            </div>
            <div class="message-content">{{ msg.content }}</div>
          </div>
        }
        @if (isStreaming) {
          <div class="message assistant streaming">
            <div class="message-header">
              <span class="role-icon"><i class="fa-regular fa-robot"></i></span>
              <span class="role-label">Claude</span>
              <span class="typing-indicator">
                <span></span><span></span><span></span>
              </span>
            </div>
            <div class="message-content">{{ streamingContent || '...' }}</div>
          </div>
        }
      </div>

      <div class="input-area">
        <textarea
          [(ngModel)]="inputText"
          (keydown.enter)="onEnterKey($event)"
          placeholder="Message Claude..."
          [disabled]="isStreaming"
          rows="1"
          #inputField>
        </textarea>
        <button
          class="send-btn"
          (click)="sendMessage()"
          [disabled]="!inputText.trim() || isStreaming">
          <i class="fa-regular fa-paper-plane"></i>
        </button>
      </div>
    </div>
  `,
  styles: [`
    .claude-panel {
      display: flex;
      flex-direction: column;
      height: 100%;
      background: #1e1e1e;
      color: #ccc;
    }
    .panel-header {
      display: flex;
      align-items: center;
      padding: 8px 12px;
      background: #333;
      border-bottom: 1px solid #444;
      gap: 8px;
    }
    .panel-title {
      flex: 1;
      font-size: 11px;
      font-weight: 600;
      color: #999;
      letter-spacing: 0.5px;
      display: flex;
      align-items: center;
      gap: 6px;
    }
    .panel-title i {
      color: #6b9eff;
    }
    .connection-status {
      font-size: 10px;
      padding: 2px 8px;
      border-radius: 10px;
      background: #5a3030;
      color: #ff8080;
    }
    .connection-status.connected {
      background: #305a30;
      color: #80ff80;
    }
    .panel-btn {
      background: none;
      border: none;
      color: #888;
      cursor: pointer;
      padding: 4px 6px;
      border-radius: 3px;
    }
    .panel-btn:hover {
      background: #444;
      color: #fff;
    }
    .messages-container {
      flex: 1;
      overflow-y: auto;
      padding: 12px;
      display: flex;
      flex-direction: column;
      gap: 12px;
    }
    .empty-state {
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      height: 100%;
      color: #666;
      text-align: center;
    }
    .empty-state i {
      font-size: 48px;
      margin-bottom: 16px;
      color: #444;
    }
    .empty-state .hint {
      font-size: 12px;
      color: #555;
    }
    .message {
      padding: 12px;
      border-radius: 8px;
      background: #2d2d2d;
    }
    .message.user {
      background: #1a3a5c;
      margin-left: 24px;
    }
    .message.assistant {
      background: #2d2d2d;
      margin-right: 24px;
    }
    .message.system {
      background: #3d3020;
      font-size: 12px;
      text-align: center;
    }
    .message-header {
      display: flex;
      align-items: center;
      gap: 8px;
      margin-bottom: 8px;
      font-size: 11px;
      color: #888;
    }
    .role-icon {
      width: 20px;
      height: 20px;
      display: flex;
      align-items: center;
      justify-content: center;
      background: #444;
      border-radius: 50%;
      font-size: 10px;
    }
    .message.user .role-icon {
      background: #2a5a8c;
    }
    .message.assistant .role-icon {
      background: #4a3a6c;
      color: #a0a0ff;
    }
    .role-label {
      font-weight: 500;
      text-transform: capitalize;
    }
    .timestamp {
      margin-left: auto;
      font-size: 10px;
    }
    .message-content {
      font-size: 13px;
      line-height: 1.5;
      white-space: pre-wrap;
      word-break: break-word;
    }
    .streaming .message-content {
      border-right: 2px solid #6b9eff;
      animation: blink 0.8s infinite;
    }
    @keyframes blink {
      0%, 50% { border-color: #6b9eff; }
      51%, 100% { border-color: transparent; }
    }
    .typing-indicator {
      display: flex;
      gap: 3px;
      margin-left: 8px;
    }
    .typing-indicator span {
      width: 4px;
      height: 4px;
      background: #6b9eff;
      border-radius: 50%;
      animation: bounce 1.2s infinite;
    }
    .typing-indicator span:nth-child(2) { animation-delay: 0.2s; }
    .typing-indicator span:nth-child(3) { animation-delay: 0.4s; }
    @keyframes bounce {
      0%, 60%, 100% { transform: translateY(0); }
      30% { transform: translateY(-4px); }
    }
    .input-area {
      display: flex;
      padding: 12px;
      background: #252526;
      border-top: 1px solid #444;
      gap: 8px;
    }
    .input-area textarea {
      flex: 1;
      background: #1e1e1e;
      border: 1px solid #444;
      border-radius: 6px;
      padding: 10px 12px;
      color: #ccc;
      font-family: inherit;
      font-size: 13px;
      resize: none;
      min-height: 40px;
      max-height: 120px;
    }
    .input-area textarea:focus {
      outline: none;
      border-color: #6b9eff;
    }
    .input-area textarea::placeholder {
      color: #666;
    }
    .send-btn {
      background: #0e639c;
      border: none;
      border-radius: 6px;
      padding: 10px 16px;
      color: #fff;
      cursor: pointer;
      font-size: 14px;
    }
    .send-btn:hover:not(:disabled) {
      background: #1177bb;
    }
    .send-btn:disabled {
      opacity: 0.5;
      cursor: not-allowed;
    }
  `]
})
export class ClaudePanelComponent implements OnInit, OnDestroy {
  @ViewChild('messagesContainer') messagesContainer!: ElementRef;
  @ViewChild('inputField') inputField!: ElementRef;

  messages: Message[] = [];
  inputText: string = '';
  isStreaming: boolean = false;
  streamingContent: string = '';
  connected: boolean = false;

  private ws: WebSocket | null = null;
  private wsUrl: string = 'ws://localhost:9877/ws';

  ngOnInit(): void {
    this.connect();
  }

  ngOnDestroy(): void {
    this.disconnect();
  }

  connect(): void {
    if (this.ws) {
      this.disconnect();
    }

    try {
      this.ws = new WebSocket(this.wsUrl);

      this.ws.onopen = () => {
        this.connected = true;
        this.addSystemMessage('Connected to Core');
        // Subscribe to claude channel for responses
        this.sendWsMessage({ type: 'subscribe', data: 'claude' });
      };

      this.ws.onmessage = (event) => {
        this.handleWsMessage(event.data);
      };

      this.ws.onclose = () => {
        this.connected = false;
        this.addSystemMessage('Disconnected from Core');
      };

      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        this.connected = false;
      };
    } catch (error) {
      console.error('Failed to connect:', error);
    }
  }

  disconnect(): void {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.connected = false;
  }

  handleWsMessage(data: string): void {
    try {
      const msg: WsMessage = JSON.parse(data);

      switch (msg.type) {
        case 'claude_response':
          this.isStreaming = false;
          this.messages.push({
            role: 'assistant',
            content: msg.data,
            timestamp: new Date()
          });
          this.scrollToBottom();
          break;

        case 'claude_stream':
          this.isStreaming = true;
          this.streamingContent = (this.streamingContent || '') + msg.data;
          this.scrollToBottom();
          break;

        case 'claude_stream_end':
          this.isStreaming = false;
          if (this.streamingContent) {
            this.messages.push({
              role: 'assistant',
              content: this.streamingContent,
              timestamp: new Date()
            });
            this.streamingContent = '';
          }
          this.scrollToBottom();
          break;

        case 'error':
          this.addSystemMessage(`Error: ${msg.data}`);
          this.isStreaming = false;
          break;
      }
    } catch (error) {
      console.error('Failed to parse message:', error);
    }
  }

  sendWsMessage(msg: any): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(msg));
    }
  }

  sendMessage(): void {
    const text = this.inputText.trim();
    if (!text) return;

    this.messages.push({
      role: 'user',
      content: text,
      timestamp: new Date()
    });

    // Send to backend via WebSocket
    this.sendWsMessage({
      type: 'claude_message',
      data: text
    });

    this.inputText = '';
    this.isStreaming = true;
    this.streamingContent = '';
    this.scrollToBottom();

    // Focus back on input
    setTimeout(() => {
      if (this.inputField) {
        this.inputField.nativeElement.focus();
      }
    }, 0);
  }

  onEnterKey(event: Event): void {
    const keyEvent = event as KeyboardEvent;
    if (!keyEvent.shiftKey) {
      event.preventDefault();
      this.sendMessage();
    }
  }

  clearMessages(): void {
    this.messages = [];
    this.streamingContent = '';
    this.isStreaming = false;
  }

  addSystemMessage(content: string): void {
    this.messages.push({
      role: 'system',
      content,
      timestamp: new Date()
    });
    this.scrollToBottom();
  }

  formatTime(date: Date): string {
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  }

  scrollToBottom(): void {
    setTimeout(() => {
      if (this.messagesContainer) {
        const el = this.messagesContainer.nativeElement;
        el.scrollTop = el.scrollHeight;
      }
    }, 0);
  }
}
