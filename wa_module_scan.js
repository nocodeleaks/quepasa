'use strict';

function safeContains(hay, needle) {
  try {
    return hay.toLowerCase().indexOf(needle.toLowerCase()) >= 0;
  } catch (e) {
    return false;
  }
}

var interesting = [
  'ssl',
  'crypto',
  'boring',
  'webrtc',
  'jingle',
  'turn',
  'stun',
  'ice',
  'udp',
  'quic'
];

var mods = Process.enumerateModules();
var modHits = [];
for (var i = 0; i < mods.length; i++) {
  var m = mods[i];
  var text = (m.name || '') + ' ' + (m.path || '');
  for (var j = 0; j < interesting.length; j++) {
    if (safeContains(text, interesting[j])) {
      modHits.push({ name: m.name, path: m.path, base: m.base.toString(), size: m.size });
      break;
    }
  }
}

send({ type: 'modules', pid: Process.id, hits: modHits });

function tryFind(moduleName, exportName) {
  try {
    var exportPtr = Process.getModuleByName(moduleName).getExportByName(exportName);
    return exportPtr.toString();
  } catch (e) {
    return null;
  }
}

var exportChecks = [
  ['ws2_32.dll', 'connect'],
  ['ws2_32.dll', 'WSAConnect'],
  ['ws2_32.dll', 'WSAIoctl'],
  ['bcrypt.dll', 'BCryptHashData'],
  ['bcrypt.dll', 'BCryptFinishHash'],
  ['crypt32.dll', 'CryptHashData'],
  ['secur32.dll', 'EncryptMessage'],
  ['secur32.dll', 'DecryptMessage'],
  ['ncrypt.dll', 'NCryptSignHash'],
  ['ncrypt.dll', 'NCryptVerifySignature']
];

var exportHits = [];
for (var k = 0; k < exportChecks.length; k++) {
  var item = exportChecks[k];
  var exportPtrStr = tryFind(item[0], item[1]);
  if (exportPtrStr) exportHits.push({ module: item[0], export: item[1], ptr: exportPtrStr });
}

send({ type: 'exports', pid: Process.id, hits: exportHits });

console.log('module scan loaded pid=' + Process.id);
