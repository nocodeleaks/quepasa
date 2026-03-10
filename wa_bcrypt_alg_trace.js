'use strict';

function isNullPtr(ptr) {
  try {
    return ptr === null || ptr === undefined || ptr.toString() === '0x0';
  } catch (e) {
    return true;
  }
}

function readUtf16(ptr) {
  try {
    if (isNullPtr(ptr)) return null;
    return Memory.readUtf16String(ptr);
  } catch (e) {
    return null;
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

function safeInt(v) {
  try {
    return v.toInt32();
  } catch (e) {
    try {
      return parseInt(v.toString(), 10) || 0;
    } catch (e2) {
      return 0;
    }
  }
}

function sendEvent(kind, obj) {
  obj.kind = kind;
  obj.pid = Process.id;
  send(obj);
}

var algs = {};
var hashes = {};
var keys = {};

function ptrKey(p) {
  try {
    return p.toString();
  } catch (e) {
    return '0x0';
  }
}

var mod = Process.getModuleByName('bcrypt.dll');

Interceptor.attach(mod.getExportByName('BCryptOpenAlgorithmProvider'), {
  onEnter: function (args) {
    this.phAlgorithm = args[0];
    this.algId = readUtf16(args[1]);
    this.impl = readUtf16(args[2]);
    this.flags = safeInt(args[3]);
  },
  onLeave: function (retval) {
    var rc = safeInt(retval);
    var algHandle = null;
    try {
      algHandle = Memory.readPointer(this.phAlgorithm);
    } catch (e) {
    }
    if (algHandle && ptrKey(algHandle) !== '0x0') {
      algs[ptrKey(algHandle)] = {
        algId: this.algId,
        impl: this.impl,
        flags: this.flags
      };
    }
    sendEvent('BCryptOpenAlgorithmProvider', {
      rc: rc,
      alg_handle: algHandle ? ptrKey(algHandle) : null,
      alg_id: this.algId,
      impl: this.impl,
      flags: this.flags
    });
  }
});

Interceptor.attach(mod.getExportByName('BCryptCreateHash'), {
  onEnter: function (args) {
    this.phHash = args[0];
    this.hAlgorithm = args[1];
    this.secret = args[4];
    this.secretLen = safeInt(args[5]);
  },
  onLeave: function (retval) {
    var rc = safeInt(retval);
    var hashHandle = null;
    try {
      hashHandle = Memory.readPointer(this.phHash);
    } catch (e) {
    }
    var secretHex = null;
    if (this.secretLen > 0 && this.secretLen <= 128) {
      secretHex = toHex(readBytes(this.secret, this.secretLen, 64));
    }
    if (hashHandle && ptrKey(hashHandle) !== '0x0') {
      hashes[ptrKey(hashHandle)] = {
        alg: algs[ptrKey(this.hAlgorithm)] || { alg_handle: ptrKey(this.hAlgorithm) },
        secret_len: this.secretLen,
        secret_hex: secretHex
      };
    }
    sendEvent('BCryptCreateHash', {
      rc: rc,
      hash_handle: hashHandle ? ptrKey(hashHandle) : null,
      alg_handle: ptrKey(this.hAlgorithm),
      alg: algs[ptrKey(this.hAlgorithm)] || null,
      secret_len: this.secretLen,
      secret_hex: secretHex
    });
  }
});

Interceptor.attach(mod.getExportByName('BCryptGenerateSymmetricKey'), {
  onEnter: function (args) {
    this.phKey = args[0];
    this.hAlgorithm = args[1];
    this.secret = args[4];
    this.secretLen = safeInt(args[5]);
  },
  onLeave: function (retval) {
    var rc = safeInt(retval);
    var keyHandle = null;
    try {
      keyHandle = Memory.readPointer(this.phKey);
    } catch (e) {
    }
    var secretHex = null;
    if (this.secretLen > 0 && this.secretLen <= 128) {
      secretHex = toHex(readBytes(this.secret, this.secretLen, 64));
    }
    if (keyHandle && ptrKey(keyHandle) !== '0x0') {
      keys[ptrKey(keyHandle)] = {
        alg: algs[ptrKey(this.hAlgorithm)] || { alg_handle: ptrKey(this.hAlgorithm) },
        secret_len: this.secretLen,
        secret_hex: secretHex
      };
    }
    sendEvent('BCryptGenerateSymmetricKey', {
      rc: rc,
      key_handle: keyHandle ? ptrKey(keyHandle) : null,
      alg_handle: ptrKey(this.hAlgorithm),
      alg: algs[ptrKey(this.hAlgorithm)] || null,
      secret_len: this.secretLen,
      secret_hex: secretHex
    });
  }
});

Interceptor.attach(mod.getExportByName('BCryptHashData'), {
  onEnter: function (args) {
    var hHash = ptrKey(args[0]);
    var cbInput = safeInt(args[2]);
    if (cbInput < 16 || cbInput > 512) return;
    sendEvent('BCryptHashData', {
      hash_handle: hHash,
      hash_meta: hashes[hHash] || null,
      len: cbInput,
      sample_hex: toHex(readBytes(args[1], cbInput, 64))
    });
  }
});

Interceptor.attach(mod.getExportByName('BCryptFinishHash'), {
  onEnter: function (args) {
    this.hHash = ptrKey(args[0]);
    this.pbOutput = args[1];
    this.cbOutput = safeInt(args[2]);
  },
  onLeave: function (retval) {
    sendEvent('BCryptFinishHash', {
      rc: safeInt(retval),
      hash_handle: this.hHash,
      hash_meta: hashes[this.hHash] || null,
      output_len: this.cbOutput,
      output_hex: toHex(readBytes(this.pbOutput, this.cbOutput, 64))
    });
  }
});

console.log('bcrypt alg tracer loaded pid=' + Process.id);
