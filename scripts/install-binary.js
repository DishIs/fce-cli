#!/usr/bin/env node
const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

const OWNER = 'DishIs';
const REPO = 'fce-cli';
const BINARY_NAME = 'fce';

const os = require('os');
const platform = os.platform();
const arch = os.arch();

const getPlatform = () => {
  if (platform === 'win32') return 'windows';
  if (platform === 'darwin') return 'darwin';
  return platform;
};

const getArch = () => {
  if (arch === 'x64') return 'amd64';
  if (arch === 'arm64') return 'arm64';
  return arch;
};

const OS = getPlatform();
const ARCH = getArch();

console.log(`Installing ${BINARY_NAME} for ${OS}/${ARCH}...`);

const BIN_DIR = process.env.BIN_DIR || path.join(process.env.PREFIX || '/usr/local', 'bin');
const dir = path.dirname(require.main.filename);

function install() {
  const dir = path.dirname(require.main.filename);
  const dest = path.join(BIN_DIR, BINARY_NAME);
  
  console.log(`Downloading from GitHub releases...`);
  try {
    const { execSync } = require('child_process');
    const version = execSync(`curl -fsSL https://api.github.com/repos/${OWNER}/${REPO}/releases/latest | grep -o '"tag_name": "[^"]*"' | cut -d'"' -f4`, { encoding: 'utf8' }).trim();
    
    const ext = OS === 'windows' ? 'zip' : 'tar.gz';
    const url = `https://github.com/${OWNER}/${REPO}/releases/download/${version}/fce_${version.replace('v','')}_${OS}_${ARCH}.${ext}`;
    
    execSync(`curl -fsSL "${url}" -o /tmp/fce.tar.gz && cd /tmp && tar -xzf fce.tar.gz ${BINARY_NAME} && mv ${BINARY_NAME} "${dest}" && chmod +x "${dest}"`, { stdio: 'inherit' });
    console.log(`Successfully installed to ${dest}`);
  } catch (err) {
    console.error('Failed to install:', err.message);
    process.exit(1);
  }
}

install();