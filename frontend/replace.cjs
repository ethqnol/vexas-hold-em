const fs = require('fs');
const path = require('path');

const dir = path.join(__dirname, 'src');

function traverse(currentDir) {
    const files = fs.readdirSync(currentDir);
    for (const file of files) {
        const fullPath = path.join(currentDir, file);
        if (fs.statSync(fullPath).isDirectory()) {
            traverse(fullPath);
        } else if (fullPath.endsWith('.tsx') || fullPath.endsWith('.ts')) {
            let content = fs.readFileSync(fullPath, 'utf8');
            if (content.includes('http://localhost:8080')) {
                let newContent = content.replace(/'http:\/\/localhost:8080([^']*)'/g, '`${import.meta.env.VITE_API_URL}$1`');
                newContent = newContent.replace(/"http:\/\/localhost:8080([^"]*)"/g, '`${import.meta.env.VITE_API_URL}$1`');
                newContent = newContent.replace(/`http:\/\/localhost:8080([^`]*)`/g, '`${import.meta.env.VITE_API_URL}$1`');
                
                fs.writeFileSync(fullPath, newContent, 'utf8');
                console.log(`Updated ${fullPath}`);
            }
        }
    }
}

traverse(dir);
