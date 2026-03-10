'use strict';

var hashStates = {};
var nextId = 1;

function isNullPtr(ptr) {
  try {
    return ptr === null || ptr === undefined || ptr.toString() === '0x0';
  } catch (e) {
    return true;
  }
}

function readBytes(ptr, len, maxLen) {
  var out = [];
  try {
    if (isNullPtr(ptr)) return out;
    var n = len;
    if (n < 0) n = 0;
    if (maxLen && n > maxLen) n = maxLen;
    for (var i = 0; i < n; i++) out.push(Memory.readU8(ptr.add(i)));
  } catch (e) {
    return [];
  }
  return out;
}

function toHex(bytes) {
  var s = '';
  for (var i = 0; i < bytes.length; i++) {
    var h = bytes[i].toString(16);
    if (h.length < 2) h = '0' + h;
    s += h;
  }
  return s;
}

function sendEvent(kind, obj) {
  obj.kind = kind;
  obj.pid = Process.id;
  send(obj);
}

function keyOf(ptr) {
  try {
    return ptr.toString();
  } catch (e) {
    return 'unknown';
  }
}

function getState(ptr) {
  var k = keyOf(ptr);
  if (!hashStates[k]) {
    hashStates[k] = { id: nextId++, handle: k, updates: 0, total_len: 0, chunks: [] };
  }
  return hashStates[k];
}

function scoreState(st) {
  if (!st) return false;
  if (st.total_len >= 20 && st.total_len <= 512) return true;
  if (st.chunks.length > 0 && st.chunks.length <= 8) return true;
  return false;
}

var hashDataPtr = Process.getModuleByName('bcrypt.dll').getExportByName('BCryptHashData');
Interceptor.attach(hashDataPtr, {
  onEnter: function (args) {
    try {
      var hHash = args[0];
      var pbInput = args[1];
      var cbInput = args[2].toInt32();
      var st = getState(hHash);
      var sample = readBytes(pbInput, cbInput, 64);
      st.updates += 1;
      st.total_len += cbInput;
      if (st.chunks.length < 8) {
        st.chunks.push({ len: cbInput, hex: toHex(sample) });
      }
      if (cbInput >= 20 && cbInput <= 512) {
        sendEvent('BCryptHashData', {
          hash_id: st.id,
          handle: st.handle,
          len: cbInput,
          sample_hex: toHex(sample)
        });
      }
    } catch (e) {
    }
  }
});

var finishPtr = Process.getModuleByName('bcrypt.dll').getExportByName('BCryptFinishHash');
Interceptor.attach(finishPtr, {
  onEnter: function (args) {
    this.hHash = args[0];
    this.pbOutput = args[1];
    this.cbOutput = args[2].toInt32();
  },
  onLeave: function (retval) {
    try {
      var st = getState(this.hHash);
      var out = readBytes(this.pbOutput, this.cbOutput, 64);
      if (scoreState(st)) {
        sendEvent('BCryptFinishHash', {
          hash_id: st.id,
          handle: st.handle,
          rc: retval.toInt32 ? retval.toInt32() : 0,
          output_len: this.cbOutput,
          output_hex: toHex(out),
          updates: st.updates,
          total_len: st.total_len,
          chunks: st.chunks
        });
      }
      delete hashStates[st.handle];
    } catch (e) {
    }
  }
});

console.log('bcrypt tracer loaded pid=' + Process.id);
