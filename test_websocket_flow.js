#!/usr/bin/env node

const WebSocket = require('ws');

let token = null;
let messageCount = 0;

function log(message) {
    console.log(`[${new Date().toISOString()}] ${message}`);
}

function connect(name) {
    return new Promise((resolve, reject) => {
        const url = token ? `ws://localhost:8080/ws?token=${token}` : 'ws://localhost:8080/ws';
        log(`${name}: Connecting to ${url}`);
        
        const ws = new WebSocket(url);
        
        ws.on('open', () => {
            log(`${name}: Connected`);
            resolve(ws);
        });
        
        ws.on('error', (error) => {
            log(`${name}: Error - ${error.message}`);
            reject(error);
        });
        
        ws.on('close', () => {
            log(`${name}: Connection closed`);
        });
        
        ws.on('message', (data) => {
            messageCount++;
            log(`${name}: ← Message #${messageCount}: ${data}`);
            
            try {
                const msg = JSON.parse(data.toString());
                
                // Save token if provided
                if ((msg.type === 'session_created' || msg.type === 'session_restored') && msg.payload) {
                    const payload = msg.payload; // Already parsed as object
                    if (payload.token) {
                        token = payload.token;
                        log(`${name}: Token saved: ${token}`);
                    }
                }
                
                // Parse hand_started message
                if (msg.type === 'hand_started') {
                    log(`${name}: ══════════════════════════════════════`);
                    log(`${name}: HAND_STARTED MESSAGE RECEIVED!`);
                    log(`${name}: Payload type: ${typeof msg.payload}`);
                    log(`${name}: Payload value: ${JSON.stringify(msg.payload)}`);
                    
                    if (msg.payload) {
                        const payload = msg.payload; // Already parsed as object
                        log(`${name}: Payload: ${JSON.stringify(payload)}`);
                        log(`${name}:   - dealerSeat: ${payload.dealerSeat}`);
                        log(`${name}:   - smallBlindSeat: ${payload.smallBlindSeat}`);
                        log(`${name}:   - bigBlindSeat: ${payload.bigBlindSeat}`);
                    }
                    log(`${name}: ══════════════════════════════════════`);
                }
                
                // Parse blind_posted message
                if (msg.type === 'blind_posted') {
                    if (msg.payload) {
                        const payload = msg.payload; // Already parsed as object
                        log(`${name}: BLIND_POSTED: seat ${payload.seatIndex}, amount ${payload.amount}, newStack ${payload.newStack}`);
                    }
                }
                
                // Parse cards_dealt message
                if (msg.type === 'cards_dealt') {
                    if (msg.payload) {
                        const payload = msg.payload; // Already parsed as object
                        log(`${name}: CARDS_DEALT: ${JSON.stringify(payload.holeCards)}`);
                    }
                }
            } catch (e) {
                log(`${name}: Parse error - ${e.message}`);
            }
        });
    });
}

function send(ws, name, message) {
    log(`${name}: → Sending: ${JSON.stringify(message)}`);
    ws.send(JSON.stringify(message));
}

function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

async function main() {
    try {
        // Create two players
        log('========================================');
        log('Starting Test: Two Players Starting Hand');
        log('========================================');
        
        const ws1 = await connect('Player1');
        await sleep(500);
        
        const ws2 = await connect('Player2');
        await sleep(500);
        
        // Set names
        send(ws1, 'Player1', { type: 'set_name', payload: { name: 'Alice' } });
        await sleep(300);
        
        send(ws2, 'Player2', { type: 'set_name', payload: { name: 'Bob' } });
        await sleep(300);
        
        // Join table
        send(ws1, 'Player1', { type: 'join_table', payload: { tableId: 'table-1' } });
        await sleep(300);
        
        send(ws2, 'Player2', { type: 'join_table', payload: { tableId: 'table-1' } });
        await sleep(500);
        
        // Start hand
        log('========================================');
        log('STARTING HAND NOW...');
        log('========================================');
        send(ws1, 'Player1', { type: 'start_hand', payload: {} });
        
        // Wait for messages
        await sleep(2000);
        
        log('========================================');
        log(`Test complete. Total messages received: ${messageCount}`);
        log('========================================');
        
        // Cleanup
        ws1.close();
        ws2.close();
        
        await sleep(500);
        process.exit(0);
    } catch (error) {
        log(`Error: ${error.message}`);
        process.exit(1);
    }
}

main();
