#!/usr/bin/env node
/**
 * MCP Gateway - Unified MCP Server Orchestrator
 * 
 * Provides:
 * 1. HTTP/SSE endpoints for all MCPs (consistent interface)
 * 2. Dynamic spawning of npm-based MCPs in isolated containers
 * 3. Health monitoring and auto-restart
 * 4. Single configuration point for all IDEs
 */

const express = require('express');
const { spawn } = require('child_process');
const cors = require('cors');
const path = require('path');
const fs = require('fs');

const app = express();
app.use(cors());
app.use(express.json());

// Configuration
const PORT = process.env.PORT || 3000;
const MCP_REGISTRY = process.env.MCP_REGISTRY || '/app/mcp-registry.json';
const DATA_DIR = process.env.MCP_DATA_DIR || '/data';

// State
const mcpInstances = new Map();
const mcpRegistry = loadRegistry();

function loadRegistry() {
    if (fs.existsSync(MCP_REGISTRY)) {
        return JSON.parse(fs.readFileSync(MCP_REGISTRY, 'utf8'));
    }
    return { mcps: {} };
}

// ── Health Check ─────────────────────────────────────────────
app.get('/health', (req, res) => {
    res.json({ 
        status: 'ok', 
        gateway: true,
        mcps: Array.from(mcpInstances.entries()).map(([name, info]) => ({
            name,
            status: info.process ? 'running' : 'stopped',
            pid: info.process?.pid
        }))
    });
});

// ── List all MCPs ────────────────────────────────────────────
app.get('/mcps', (req, res) => {
    const allMcps = Object.entries(mcpRegistry.mcps).map(([name, config]) => {
        const instance = mcpInstances.get(name);
        return {
            name,
            type: config.type,
            status: instance?.process ? 'running' : (config.type === 'http' ? 'available' : 'stopped'),
            endpoint: instance ? `/mcp/${name}/sse` : null,
            externalUrl: config.type === 'http' ? config.url : null,
            config: { 
                description: config.description,
                package: config.package 
            }
        };
    });
    res.json(allMcps);
});

// ── Start an MCP ─────────────────────────────────────────────
app.post('/mcps/:name/start', async (req, res) => {
    const name = req.params.name;
    const config = mcpRegistry.mcps[name];
    
    if (!config) {
        return res.status(404).json({ error: `MCP ${name} not found in registry` });
    }
    
    if (mcpInstances.has(name)) {
        return res.json({ 
            name, 
            status: 'already_running',
            endpoint: `/mcp/${name}/sse`
        });
    }
    
    try {
        const instance = await startMcp(name, config);
        mcpInstances.set(name, instance);

        const resp = {
            name,
            status: 'started',
            endpoint: `/mcp/${name}/sse`
        };

        // Include PID when the MCP has a process (npm/docker). For HTTP-type MCPs expose the URL.
        if (instance && instance.process && instance.process.pid) {
            resp.pid = instance.process.pid;
        } else if (instance && instance.type === 'http' && instance.url) {
            resp.url = instance.url;
            resp.mode = 'http_proxy';
        }

        res.json(resp);
    } catch (err) {
        res.status(500).json({ error: err.message });
    }
});

// ── Stop an MCP ──────────────────────────────────────────────
app.post('/mcps/:name/stop', (req, res) => {
    const name = req.params.name;
    const instance = mcpInstances.get(name);
    
    if (!instance) {
        return res.status(404).json({ error: `MCP ${name} not running` });
    }
    
    stopMcp(name, instance);
    res.json({ name, status: 'stopped' });
});

// ── SSE Endpoint for MCP communication ───────────────────────
app.get('/mcp/:name/sse', (req, res) => {
    const name = req.params.name;
    const instance = mcpInstances.get(name);
    const config = mcpRegistry.mcps[name];
    
    if (!instance) {
        return res.status(404).json({ error: `MCP ${name} not running. Start with POST /mcps/${name}/start` });
    }
    
    // Handle HTTP-type MCPs (external URLs)
    if (instance.type === 'http' || config?.type === 'http') {
        // For HTTP MCPs, redirect to external URL
        const targetUrl = instance.url || config.url;
        res.writeHead(200, {
            'Content-Type': 'text/event-stream',
            'Cache-Control': 'no-cache',
            'Connection': 'keep-alive'
        });
        
        // Return connection info - clients should use the external URL directly
        res.write(`data: ${JSON.stringify({ type: 'info', url: targetUrl, mode: 'http_proxy' })}

`);
        
        // Keep connection alive
        const keepAlive = setInterval(() => {
            res.write(`:ping\n\n`);
        }, 30000);
        
        req.on('close', () => {
            clearInterval(keepAlive);
        });
        return;
    }
    
    // SSE setup
    res.writeHead(200, {
        'Content-Type': 'text/event-stream',
        'Cache-Control': 'no-cache',
        'Connection': 'keep-alive'
    });
    
    // Forward messages from MCP stdout to SSE
    instance.process.stdout.on('data', (data) => {
        const lines = data.toString().split('\n').filter(l => l.trim());
        for (const line of lines) {
            try {
                const msg = JSON.parse(line);
                res.write(`data: ${JSON.stringify(msg)}\n\n`);
            } catch (e) {
                // Non-JSON output, log but don't crash
                console.error(`[${name}]`, line);
            }
        }
    });
    
    // Handle client disconnect
    req.on('close', () => {
        // Don't stop the MCP, just disconnect this client
    });
});

// ── POST endpoint for MCP messages ─────────────────────────
app.post('/mcp/:name/message', express.json(), (req, res) => {
    const name = req.params.name;
    const instance = mcpInstances.get(name);
    const config = mcpRegistry.mcps[name];
    
    if (!instance) {
        return res.status(404).json({ error: `MCP ${name} not running` });
    }
    
    // Handle HTTP-type MCPs
    if (instance.type === 'http' || config?.type === 'http') {
        // For HTTP MCPs, proxy the message to external URL
        const targetUrl = instance.url || config.url;
        // Return 200 with the target URL for clients to use directly
        return res.json({ 
            type: 'http_proxy', 
            url: targetUrl,
            message: 'Use this URL directly for HTTP MCP connections'
        });
    }
    
    const message = req.body;
    instance.process.stdin.write(JSON.stringify(message) + '\n');
    
    // For simplicity, return 202 Accepted
    res.status(202).json({ status: 'sent' });
});

// ── Start MCP Helper ─────────────────────────────────────────
async function startMcp(name, config) {
    let childProcess;
    const envVars = { ...process.env, ...(config.env || {}) };
    
    if (config.type === 'npm') {
        // Run npm-based MCP
        const args = ['-y', config.package];
        if (config.args) {
            args.push(...config.args);
        }
        
        childProcess = spawn('npx', args, {
            stdio: ['pipe', 'pipe', 'pipe'],
            env: envVars
        });
    } else if (config.type === 'docker') {
        // Run Docker-based MCP
        const args = ['run', '-i', '--rm'];
        if (config.volumes) {
            for (const vol of config.volumes) {
                args.push('-v', vol);
            }
        }
        args.push(config.image);
        if (config.args) {
            args.push(...config.args);
        }
        
        childProcess = spawn('docker', args, {
            stdio: ['pipe', 'pipe', 'pipe']
        });
    } else if (config.type === 'http') {
        // HTTP-based MCP - no process, just proxy to external URL
        return { 
            type: 'http', 
            url: config.url, 
            config 
        };
    } else {
        throw new Error(`Unknown MCP type: ${config.type}`);
    }
    
    // Handle errors
    childProcess.stderr.on('data', (data) => {
        console.error(`[${name}]`, data.toString());
    });
    
    childProcess.on('exit', (code) => {
        console.log(`[${name}] exited with code ${code}`);
        mcpInstances.delete(name);
    });
    
    // Wait a moment for startup
    await new Promise(resolve => setTimeout(resolve, 2000));
    
    return { process: childProcess, config };
}

// ── Stop MCP Helper ──────────────────────────────────────────
function stopMcp(name, instance) {
    if (instance.type === 'http') {
        // HTTP MCPs have no process to kill, just remove from registry
        mcpInstances.delete(name);
        return;
    }
    
    instance.process.kill('SIGTERM');
    
    // Force kill after 5 seconds
    setTimeout(() => {
        if (!instance.process.killed) {
            instance.process.kill('SIGKILL');
        }
    }, 5000);
    
    mcpInstances.delete(name);
}

// ── Auto-start required MCPs ─────────────────────────────────
async function autoStartMcps() {
    for (const [name, config] of Object.entries(mcpRegistry.mcps)) {
        if (config.required) {
            console.log(`[gateway] Auto-starting required MCP: ${name}`);
            try {
                const instance = await startMcp(name, config);
                mcpInstances.set(name, instance);
            } catch (err) {
                console.error(`[gateway] Failed to start ${name}:`, err.message);
            }
        }
    }
}

// ── Start Server ─────────────────────────────────────────────
app.listen(PORT, '0.0.0.0', async () => {
    console.log(`[gateway] MCP Gateway running on port ${PORT}`);
    console.log(`[gateway] Registry: ${MCP_REGISTRY}`);
    console.log(`[gateway] MCPs configured: ${Object.keys(mcpRegistry.mcps).length}`);
    
    await autoStartMcps();
});
