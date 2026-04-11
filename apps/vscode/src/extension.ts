import * as vscode from 'vscode';

// Brain daemon client
class BrainDaemonClient {
  private daemonUrl: string;

  constructor() {
    const config = vscode.workspace.getConfiguration('brain');
    this.daemonUrl = config.get('daemonUrl', 'http://localhost:8080');
  }

  async getArtifacts(): Promise<any[]> {
    // In real implementation: HTTP GET to ${this.daemonUrl}/api/skills etc.
    return [];
  }

  async getContextBundle(): Promise<any> {
    return null;
  }

  async getPolicy(): Promise<any> {
    return null;
  }
}

// View providers
class ArtifactViewProvider implements vscode.TreeDataProvider<ArtifactItem> {
  private _onDidChangeTreeData = new vscode.EventEmitter<ArtifactItem | undefined>();
  readonly onDidChangeTreeData = this._onDidChangeTreeData.event;

  constructor(private client: BrainDaemonClient) {}

  getTreeItem(element: ArtifactItem): vscode.TreeItem {
    return element;
  }

  async getChildren(element?: ArtifactItem): Promise<ArtifactItem[]> {
    if (!element) {
      return [
        new ArtifactItem('Skills', vscode.TreeItemCollapsibleState.Collapsed),
        new ArtifactItem('MCPs', vscode.TreeItemCollapsibleState.Collapsed),
        new ArtifactItem('Agents', vscode.TreeItemCollapsibleState.Collapsed),
      ];
    }
    return [];
  }

  refresh(): void {
    this._onDidChangeTreeData.fire(undefined);
  }
}

class ArtifactItem extends vscode.TreeItem {
  constructor(label: string, collapsibleState: vscode.TreeItemCollapsibleState) {
    super(label, collapsibleState);
    this.tooltip = label;
  }
}

// Activation
export function activate(context: vscode.ExtensionContext) {
  const client = new BrainDaemonClient();
  
  // Register view provider
  const artifactProvider = new ArtifactViewProvider(client);
  vscode.window.registerTreeDataProvider('brain.artifacts', artifactProvider);

  // Register commands
  context.subscriptions.push(
    vscode.commands.registerCommand('brain.showArtifacts', () => {
      vscode.window.showInformationMessage('Brain Artifacts panel opened');
    }),
    vscode.commands.registerCommand('brain.showContext', () => {
      vscode.window.showInformationMessage('Brain Context Bundle');
    }),
    vscode.commands.registerCommand('brain.showPolicy', () => {
      vscode.window.showInformationMessage('Brain Policy Resolver');
    }),
    vscode.commands.registerCommand('brain.applySkill', async () => {
      vscode.window.showInformationMessage('Skill application triggered');
    }),
    vscode.commands.registerCommand('brain.runWorkflow', async () => {
      vscode.window.showInformationMessage('Workflow execution triggered');
    })
  );

  console.log('Brain extension activated');
}

export function deactivate() {}
