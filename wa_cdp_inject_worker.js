;(function () {
  if (globalThis.__WA_WORKER_CRYPTO_MON_V1) return;
  globalThis.__WA_WORKER_CRYPTO_MON_V1 = true;

  function j(v) { try { return JSON.stringify(v); } catch (_) { return '"<unserializable>"'; } }
  function bufInfo(v) {
    try {
      if (v == null) return null;
      var u8;
      if (v instanceof ArrayBuffer) u8 = new Uint8Array(v);
      else if (ArrayBuffer.isView(v)) u8 = new Uint8Array(v.buffer, v.byteOffset, v.byteLength);
      else return { type: typeof v };
      var n = u8.byteLength;
      var lim = Math.min(n, 96);
      var hex = '';
      for (var i = 0; i < lim; i++) { var h = u8[i].toString(16); if (h.length < 2) h = '0' + h; hex += h; }
      var info = { len: n, head_hex: hex };
      if (n <= 256) {
        var full = '';
        for (var k = 0; k < n; k++) { var fh = u8[k].toString(16); if (fh.length < 2) fh = '0' + fh; full += fh; }
        info.full_hex = full;
      } else {
        var tail = '';
        var start = Math.max(0, n - 96);
        for (var j = start; j < n; j++) { var th = u8[j].toString(16); if (th.length < 2) th = '0' + th; tail += th; }
        info.tail_hex = tail;
      }
      return info;
    } catch (e) {
      return { err: String(e) };
    }
  }
  function algInfo(a) {
    try {
      if (!a) return null;
      if (typeof a === 'string') return { name: a };
      var o = {};
      Object.keys(a).forEach(function (k) {
        var v = a[k];
        if (typeof v === 'string' || typeof v === 'number' || typeof v === 'boolean') o[k] = v;
        else if (v instanceof ArrayBuffer || ArrayBuffer.isView(v)) o[k] = bufInfo(v);
      });
      return o;
    } catch (e) {
      return { err: String(e) };
    }
  }
  function stackStr() {
    try {
      return String(new Error().stack || '').split('\n').slice(2, 8).join(' | ');
    } catch (_) {
      return '';
    }
  }

  var subtle = globalThis.crypto && globalThis.crypto.subtle;
  if (!subtle) {
    try { console.log('[WA-WORKER] crypto.missing {}'); } catch (_) {}
    return;
  }

  var proto = globalThis.SubtleCrypto && globalThis.SubtleCrypto.prototype;
  if (proto) {
    ['encrypt', 'decrypt', 'importKey'].forEach(function (name) {
      var orig = proto[name];
      if (typeof orig !== 'function' || orig.__waWorkerProtoWrappedV1) return;
      function wrapped() {
        var args = Array.prototype.slice.call(arguments);
        var payload = { method: name, stack: stackStr() };
        try {
          if (name === 'importKey') {
            payload.format = args[0];
            payload.keyData = bufInfo(args[1]);
            payload.algorithm = algInfo(args[2]);
            payload.extractable = !!args[3];
            payload.usages = args[4];
          } else {
            payload.algorithm = algInfo(args[0]);
            payload.keyType = args[1] && args[1].type ? args[1].type : typeof args[1];
            payload.data = bufInfo(args[2]);
          }
        } catch (e) {
          payload.err = String(e);
        }
        try { console.log('[WA-WORKER] crypto.proto.' + name + ' ' + j(payload)); } catch (_) {}
        return orig.apply(this, arguments);
      }
      wrapped.__waWorkerProtoWrappedV1 = true;
      proto[name] = wrapped;
    });
  }

  ['encrypt', 'decrypt'].forEach(function (name) {
    var orig = subtle[name];
    if (typeof orig !== 'function' || orig.__waWorkerWrappedV1) return;
    function wrapped() {
      var args = Array.prototype.slice.call(arguments);
      var payload = { method: name, stack: stackStr() };
      try {
        payload.algorithm = algInfo(args[0]);
        payload.keyType = args[1] && args[1].type ? args[1].type : typeof args[1];
        payload.data = bufInfo(args[2]);
      } catch (e) {
        payload.err = String(e);
      }
      try { console.log('[WA-WORKER] crypto.direct.' + name + ' ' + j(payload)); } catch (_) {}
      var ret = orig.apply(subtle, arguments);
      try {
        if (ret && typeof ret.then === 'function') {
          ret.then(function (value) {
            try { console.log('[WA-WORKER] crypto.result.' + name + ' ' + j(bufInfo(value))); } catch (_) {}
            return value;
          });
        }
      } catch (_) {}
      return ret;
    }
    wrapped.__waWorkerWrappedV1 = true;
    subtle[name] = wrapped;
  });

  try { console.log('[WA-WORKER] installed ' + j({ ok: true })); } catch (_) {}
})();
