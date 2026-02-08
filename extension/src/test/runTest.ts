import * as path from 'path';
import * as os from 'os';
import * as fs from 'fs';
import { runTests } from '@vscode/test-electron';

async function main() {
    try {
        // Inject a shim 'cursor-rules' into PATH so commands the extension calls succeed.
        const tmpBin = fs.mkdtempSync(path.join(os.tmpdir(), 'cursor-rules-shim-'));
        const shimPath = path.join(tmpBin, process.platform === 'win32' ? 'cursor-rules.cmd' : 'cursor-rules');
        const shimScript = process.platform === 'win32'
            ? '@echo off\r\n' +
            'echo Package dir: C:\\tmp\\cursor-rules\r\n' +
            'echo - frontend.mdc\r\n'
            : '#!/usr/bin/env bash\n' +
            'set -e\n' +
            'case "$1" in\n' +
            '  sync) echo "Package dir: /tmp/cursor-rules"; echo "- frontend.mdc";;\n' +
            '  effective) echo "---\\n# frontend.mdc\\n@file /tmp/cursor-rules/frontend.mdc";;\n' +
            '  install) echo "Installed preset \"$2\"";;\n' +
            '  *) echo "ok";;\n' +
            'esac\n';
        fs.writeFileSync(shimPath, shimScript, { mode: 0o755 });
        process.env.PATH = `${tmpBin}${path.delimiter}${process.env.PATH ?? ''}`;

        const extensionDevelopmentPath = path.resolve(__dirname, '../../');
        const extensionTestsPath = path.resolve(__dirname, './suite/index');

        await runTests({ extensionDevelopmentPath, extensionTestsPath });
    } catch (err) {
        console.error('Failed to run tests');
        process.exit(1);
    }
}

main();


