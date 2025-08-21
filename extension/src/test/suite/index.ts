import * as path from 'path';
import * as Mocha from 'mocha';
import { glob } from 'glob';

export function run(): Promise<void> {
    const mocha = new Mocha({ ui: 'bdd', color: true, timeout: 10000 });
    const testsRoot = path.resolve(__dirname, '.');

    return (async () => {
        const files: string[] = await glob('**/**.test.js', { cwd: testsRoot });
        files.forEach((f: string) => mocha.addFile(path.resolve(testsRoot, f)));

        return new Promise<void>((resolve, reject) => {
            try {
                mocha.run((failures: number) => {
                    if (failures > 0) {
                        reject(new Error(`${failures} tests failed.`));
                    } else {
                        resolve();
                    }
                });
            } catch (err) {
                reject(err);
            }
        });
    })();
}


