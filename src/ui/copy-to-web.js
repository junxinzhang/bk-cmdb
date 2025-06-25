#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

/**
 * Copy build output to local web directory
 * This script copies files from src/bin/enterprise/cmdb/web to src/web
 */

const sourceDir = path.resolve(__dirname, '../bin/enterprise/cmdb/web');
const destDir = path.resolve(__dirname, '../web');

// Utility function to copy directory recursively
function copyDirRecursive(src, dest) {
  // Create destination directory if it doesn't exist
  if (!fs.existsSync(dest)) {
    fs.mkdirSync(dest, { recursive: true });
  }

  const entries = fs.readdirSync(src, { withFileTypes: true });

  for (let entry of entries) {
    const srcPath = path.join(src, entry.name);
    const destPath = path.join(dest, entry.name);

    if (entry.isDirectory()) {
      copyDirRecursive(srcPath, destPath);
    } else {
      fs.copyFileSync(srcPath, destPath);
    }
  }
}

// Main execution
try {
  console.log('Copying built assets from enterprise build to local web directory...');
  console.log(`Source: ${sourceDir}`);
  console.log(`Destination: ${destDir}`);

  // Check if source directory exists
  if (!fs.existsSync(sourceDir)) {
    console.error(`Error: Source directory ${sourceDir} does not exist!`);
    console.error('Please run "yarn build" first to generate the build output.');
    process.exit(1);
  }

  // Remove existing destination directory contents
  if (fs.existsSync(destDir)) {
    fs.rmSync(destDir, { recursive: true, force: true });
  }

  // Copy files
  copyDirRecursive(sourceDir, destDir);

  console.log('✅ Successfully copied built assets to local web directory!');
  
  // List copied files
  const files = fs.readdirSync(destDir);
  console.log('\nCopied files and directories:');
  files.forEach(file => {
    const filePath = path.join(destDir, file);
    const stats = fs.statSync(filePath);
    const type = stats.isDirectory() ? '[DIR]' : '[FILE]';
    console.log(`  ${type} ${file}`);
  });

} catch (error) {
  console.error('❌ Error during copy operation:', error.message);
  process.exit(1);
}