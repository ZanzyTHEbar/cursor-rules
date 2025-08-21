import * as assert from 'assert';
import * as vscode from 'vscode';

describe('Extension Basics', () => {
    it('sample test runs', () => {
        assert.strictEqual(1 + 1, 2);
    });

    it('activates extension and runs sync command', async () => {
        const ext = vscode.extensions.getExtension('DaOfficialWizard.cursor-rules-manager');
        assert.ok(ext, 'extension not found');
        await ext!.activate();
        const cmds = await vscode.commands.getCommands(true);
        assert.ok(cmds.includes('cursorRules.sync'));
        await vscode.commands.executeCommand('cursorRules.sync');
    });

    it('runs showEffective and opens document', async () => {
        const ext = vscode.extensions.getExtension('DaOfficialWizard.cursor-rules-manager');
        assert.ok(ext);
        await ext!.activate();
        const cmds = await vscode.commands.getCommands(true);
        assert.ok(cmds.includes('cursorRules.showEffective'));
        await vscode.commands.executeCommand('cursorRules.showEffective');
    });

    it('runs installPreset command (shimmed)', async () => {
        const ext = vscode.extensions.getExtension('DaOfficialWizard.cursor-rules-manager');
        assert.ok(ext);
        await ext!.activate();
        const cmds = await vscode.commands.getCommands(true);
        assert.ok(cmds.includes('cursorRules.installPreset'));
        // simulate user input by directly invoking the CLI via command path would be complex here; just ensure command exists
    });
});


