import * as vscode from 'vscode';
import { exec } from 'child_process';

function runCmd(command: string, cwd?: string): Promise<string> {
    return new Promise((resolve, reject) => {
        exec(command, { cwd }, (err, stdout, stderr) => {
            if (err) return reject(stderr || err.message);
            resolve(stdout);
        });
    });
}

// Sample documents concurrently and return a map of languageId -> count
async function detectLanguagesSample(uris: vscode.Uri[], concurrency = 8): Promise<Record<string, number>> {
    const counts: Record<string, number> = {};
    for (let i = 0; i < uris.length; i += concurrency) {
        const chunk = uris.slice(i, i + concurrency);
        const promises = chunk.map(async (uri) => {
            try {
                const doc = await vscode.workspace.openTextDocument(uri);
                return doc.languageId || 'unknown';
            } catch (e) {
                return null;
            }
        });
        const results = await Promise.all(promises);
        for (const lang of results) {
            if (!lang) continue;
            counts[lang] = (counts[lang] || 0) + 1;
        }
    }
    return counts;
}

export function activate(context: vscode.ExtensionContext) {
    const output = vscode.window.createOutputChannel('Cursor Rules');

    async function detectAndOfferPresets() {
        const folders = vscode.workspace.workspaceFolders || [];
        if (folders.length === 0) return;

        // sample files with common extensions and detect languages using helper
        const samplePattern = '**/*.{js,ts,jsx,tsx,go,py,rs,java,cs,php,rb,swift,kt,scala,html,css}';
        const uris = await vscode.workspace.findFiles(samplePattern, '**/node_modules/**', 200);
        const langCount = await detectLanguagesSample(uris);

        const languages = Object.keys(langCount).sort((a, b) => (langCount[b] - langCount[a]));
        if (languages.length === 0) return;

        // fetch available presets from CLI
        let presets: string[] = [];
        try {
            const out = await runCmd('cursor-rules sync');
            const lines = out.split(/\r?\n/);
            for (const l of lines) {
                const trimmed = l.trim();
                if (trimmed.startsWith('-')) {
                    const name = trimmed.slice(1).trim();
                    if (name.endsWith('.mdc')) presets.push(name.replace(/\.mdc$/, ''));
                    else presets.push(name);
                }
            }
        } catch (e) {
            // ignore
        }

        const detected = languages.slice(0, 5).map(l => `${l} (${langCount[l]})`).join(', ');
        const pickMsg = `Detected languages: ${detected}. Select presets to install (or Cancel)`;
        let items: vscode.QuickPickItem[] = [];
        if (presets.length > 0) {
            items = presets.map(p => ({ label: p }));
        } else {
            items = languages.slice(0, 5).map(l => ({ label: l }));
        }

        const selections = await vscode.window.showQuickPick(items, { canPickMany: true, placeHolder: pickMsg });
        if (!selections || selections.length === 0) return;

        const workspaceRoot = vscode.workspace.workspaceFolders![0].uri.fsPath;
        for (const s of selections) {
            const preset = s.label;
            output.show(true);
            output.appendLine(`Running: cursor-rules install ${preset} (cwd=${workspaceRoot})`);
            try {
                const res = await runCmd(`cursor-rules install ${preset}`, workspaceRoot);
                output.appendLine(res || `Installed ${preset}`);
                vscode.window.showInformationMessage(`Installed preset ${preset}`);
            } catch (e) {
                vscode.window.showErrorMessage(String(e));
            }
        }
    }

    context.subscriptions.push(vscode.commands.registerCommand('cursorRules.sync', async () => {
        output.show(true);
        output.appendLine('Running: cursor-rules sync');
        try {
            const res = await runCmd('cursor-rules sync');
            output.appendLine(res);
        } catch (e) {
            vscode.window.showErrorMessage(String(e));
        }
    }));

    context.subscriptions.push(vscode.commands.registerCommand('cursorRules.showEffective', async () => {
        output.show(true);
        output.appendLine('Running: cursor-rules effective');
        try {
            const res = await runCmd('cursor-rules effective');
            const doc = await vscode.workspace.openTextDocument({ content: res, language: 'markdown' });
            vscode.window.showTextDocument(doc);
        } catch (e) {
            vscode.window.showErrorMessage(String(e));
        }
    }));

    context.subscriptions.push(vscode.commands.registerCommand('cursorRules.installPreset', async () => {
        const preset = await vscode.window.showInputBox({ prompt: 'Preset name to install' });
        if (!preset) return;
        output.show(true);
        output.appendLine(`Running: cursor-rules install ${preset}`);
        try {
            const cwd = vscode.workspace.workspaceFolders && vscode.workspace.workspaceFolders[0].uri.fsPath;
            const res = await runCmd(`cursor-rules install ${preset}`, cwd);
            output.appendLine(res);
            vscode.window.showInformationMessage(`Installed preset ${preset}`);
        } catch (e) {
            vscode.window.showErrorMessage(String(e));
        }
    }));

    detectAndOfferPresets().catch(err => output.appendLine('detect error: ' + String(err)));
}

export function deactivate() { }


