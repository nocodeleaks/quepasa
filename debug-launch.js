const { spawn } = require('child_process');
const path = require('path');

// Script para lançar o Delve sem warning via wrapper
const wrapperPath = path.join(__dirname, 'dlv-wrapper.bat');
const args = ['exec', path.join(__dirname, 'dist', 'win-quepasa-service.exe')];

console.log('Starting QuePasa with Delve wrapper...');

const child = spawn(wrapperPath, args, {
    stdio: 'inherit',
    shell: true,
    env: { ...process.env, CGO_ENABLED: '1' }
});

child.on('exit', (code) => {
    console.log(`Delve exited with code ${code}`);
    process.exit(code);
});

child.on('error', (err) => {
    console.error('Failed to start Delve:', err);
    process.exit(1);
});