"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.activate = activate;
exports.deactivate = deactivate;
const vscode = require("vscode");
// Brain daemon client
class BrainDaemonClient {
    constructor() {
        const config = vscode.workspace.getConfiguration('brain');
        this.daemonUrl = config.get('daemonUrl', 'http://localhost:8080');
    }
    async getArtifacts() {
        // In real implementation: HTTP GET to ${this.daemonUrl}/api/skills etc.
        return [];
    }
    async getContextBundle() {
        return null;
    }
    async getPolicy() {
        return null;
    }
}
// View providers
class ArtifactViewProvider {
    constructor(client) {
        this.client = client;
        this._onDidChangeTreeData = new vscode.EventEmitter();
        this.onDidChangeTreeData = this._onDidChangeTreeData.event;
    }
    getTreeItem(element) {
        return element;
    }
    async getChildren(element) {
        if (!element) {
            return [
                new ArtifactItem('Skills', vscode.TreeItemCollapsibleState.Collapsed),
                new ArtifactItem('MCPs', vscode.TreeItemCollapsibleState.Collapsed),
                new ArtifactItem('Agents', vscode.TreeItemCollapsibleState.Collapsed),
            ];
        }
        return [];
    }
    refresh() {
        this._onDidChangeTreeData.fire(undefined);
    }
}
class ArtifactItem extends vscode.TreeItem {
    constructor(label, collapsibleState) {
        super(label, collapsibleState);
        this.tooltip = label;
    }
}
// Activation
function activate(context) {
    const client = new BrainDaemonClient();
    // Register view provider
    const artifactProvider = new ArtifactViewProvider(client);
    vscode.window.registerTreeDataProvider('brain.artifacts', artifactProvider);
    // Register commands
    context.subscriptions.push(vscode.commands.registerCommand('brain.showArtifacts', () => {
        vscode.window.showInformationMessage('Brain Artifacts panel opened');
    }), vscode.commands.registerCommand('brain.showContext', () => {
        vscode.window.showInformationMessage('Brain Context Bundle');
    }), vscode.commands.registerCommand('brain.showPolicy', () => {
        vscode.window.showInformationMessage('Brain Policy Resolver');
    }), vscode.commands.registerCommand('brain.applySkill', async () => {
        vscode.window.showInformationMessage('Skill application triggered');
    }), vscode.commands.registerCommand('brain.runWorkflow', async () => {
        vscode.window.showInformationMessage('Workflow execution triggered');
    }));
    console.log('Brain extension activated');
}
function deactivate() { }
//# sourceMappingURL=extension.js.map