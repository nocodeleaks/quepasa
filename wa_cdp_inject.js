(function () {
  var WA_MON_VERSION = 8;
  var WA_MON_SEQ = 0;
  var PREV_WA_MON_VERSION = globalThis.__waMonVersion || 0;
  if (PREV_WA_MON_VERSION >= WA_MON_VERSION) {
    return;
  }
  globalThis.__waMonInstalled = true;
  globalThis.__waMonVersion = WA_MON_VERSION;
  var WRAP_TAG = "__waWrapped";
  var WRAP_CTOR_TAG = "__waWrappedCtor";

  function safe(value) {
    try {
      return JSON.stringify(value);
    } catch (_) {
      try {
        return String(value);
      } catch (_) {
        return "<unserializable>";
      }
    }
  }

  function asciiHeadFromBytes(bytes, maxLen) {
    var n = Math.min(bytes.length || 0, maxLen || 16);
    var out = "";
    for (var i = 0; i < n; i++) {
      var b = bytes[i];
      if (b >= 32 && b <= 126) {
        out += String.fromCharCode(b);
      } else {
        out += ".";
      }
    }
    return out;
  }

  function log(kind, payload) {
    try {
      WA_MON_SEQ += 1;
      var enriched = payload || {};
      enriched.ts = Date.now();
      enriched.seq = WA_MON_SEQ;
      console.log("[WA-MON]", kind, safe(enriched));
    } catch (_) {}
  }

  function looksLikeCallText(text) {
    try {
      if (typeof text !== "string" || text.length === 0) return false;
      return (
        text.indexOf("<call") >= 0 ||
        text.indexOf("<relaylatency") >= 0 ||
        text.indexOf("<transport") >= 0 ||
        text.indexOf("<accept") >= 0 ||
        text.indexOf("<preaccept") >= 0 ||
        text.indexOf("<terminate") >= 0 ||
        text.indexOf("<mute_v2") >= 0 ||
        text.indexOf("<receipt") >= 0 ||
        text.indexOf("call-id=") >= 0 ||
        text.indexOf("transport-message-type") >= 0
      );
    } catch (_) {
      return false;
    }
  }

  function logCallText(kind, text, extra) {
    try {
      if (!looksLikeCallText(text)) return;
      var payload = extra || {};
      payload.len = text.length;
      payload.head = text.slice(0, 512);
      if (text.length <= 4096) {
        payload.full = text;
      }
      log(kind, payload);
    } catch (_) {}
  }

  function shortValue(value) {
    try {
      if (value == null) return value;
      if (typeof value === "string") {
        return {
          type: "string",
          len: value.length,
          head: value.slice(0, 240)
        };
      }
      var bin = describeBinary(value);
      if (bin) return bin;
      if (Array.isArray(value)) {
        return {
          type: "array",
          len: value.length,
          sample: value.slice(0, 6).map(shortValue)
        };
      }
      if (typeof value === "object") {
        var out = {
          type: Object.prototype.toString.call(value),
          keys: []
        };
        var keys = [];
        try {
          keys = Object.keys(value).slice(0, 16);
        } catch (_) {}
        out.keys = keys;
        var picked = {};
        for (var i = 0; i < keys.length; i++) {
          var k = keys[i];
          var v;
          try {
            v = value[k];
          } catch (_) {
            continue;
          }
          if (typeof v === "string" || typeof v === "number" || typeof v === "boolean" || v == null) {
            picked[k] = v;
          } else {
            var innerBin = describeBinary(v);
            if (innerBin) picked[k] = innerBin;
          }
        }
        if (Object.keys(picked).length > 0) {
          out.sample = picked;
        }
        return out;
      }
      return {
        type: typeof value,
        value: String(value)
      };
    } catch (err) {
      return {
        err: String(err)
      };
    }
  }

  function logCallResult(kind, result) {
    try {
      if (result && typeof result.then === "function") {
        return result.then(
          function (value) {
            log(kind + ":ok", { value: shortValue(value) });
            return value;
          },
          function (err) {
            log(kind + ":err", { message: err && err.message ? err.message : String(err) });
            throw err;
          }
        );
      }
      log(kind + ":ok", { value: shortValue(result) });
      return result;
    } catch (err) {
      log(kind + ":err", { message: err && err.message ? err.message : String(err) });
      throw err;
    }
  }

  function wrapMethodOnce(holder, methodName, pathLabel) {
    try {
      if (!holder || typeof holder[methodName] !== "function") return false;
      var original = holder[methodName];
      if (original.__waCallHookWrapped) return false;
      var wrapped = function () {
        var args = [];
        for (var i = 0; i < arguments.length; i++) {
          args.push(shortValue(arguments[i]));
        }
        log("stanza." + methodName, {
          path: pathLabel,
          argc: arguments.length,
          args: args
        });
        return logCallResult("stanza." + methodName, original.apply(this, arguments));
      };
      wrapped.__waCallHookWrapped = true;
      holder[methodName] = wrapped;
      log("stanza.wrap", {
        path: pathLabel,
        method: methodName
      });
      return true;
    } catch (err) {
      log("stanza.wrap.err", {
        path: pathLabel,
        method: methodName,
        message: err && err.message ? err.message : String(err)
      });
      return false;
    }
  }

  function looksTraversable(value) {
    if (!value) return false;
    var t = typeof value;
    if (t !== "object" && t !== "function") return false;
    if (value === globalThis) return true;
    var tag = "";
    try {
      tag = Object.prototype.toString.call(value);
    } catch (_) {}
    if (tag === "[object Window]" || tag === "[object HTMLDocument]" || tag === "[object Document]" || tag === "[object Location]") {
      return false;
    }
    return true;
  }

  function scanForStanzaHooks() {
    var queue = [{ value: globalThis, path: "globalThis", depth: 0 }];
    var seen = typeof WeakSet === "function" ? new WeakSet() : null;
    var visited = 0;
    var wrapped = 0;
    while (queue.length > 0 && visited < 1200) {
      var item = queue.shift();
      var value = item.value;
      if (!looksTraversable(value)) continue;
      if (seen) {
        if (seen.has(value)) continue;
        seen.add(value);
      }
      visited += 1;

      try {
        if (typeof value.callStanza === "function") wrapped += wrapMethodOnce(value, "callStanza", item.path + ".callStanza") ? 1 : 0;
        if (typeof value.callStanzaAsync === "function") wrapped += wrapMethodOnce(value, "callStanzaAsync", item.path + ".callStanzaAsync") ? 1 : 0;
        if (typeof value.castStanza === "function") wrapped += wrapMethodOnce(value, "castStanza", item.path + ".castStanza") ? 1 : 0;
        if (typeof value.castStanzaAsync === "function") wrapped += wrapMethodOnce(value, "castStanzaAsync", item.path + ".castStanzaAsync") ? 1 : 0;
      } catch (_) {}

      if (item.depth >= 3) continue;

      var props = [];
      try {
        props = Object.getOwnPropertyNames(value).slice(0, 80);
      } catch (_) {
        continue;
      }

      for (var i = 0; i < props.length; i++) {
        var key = props[i];
        if (key === "window" || key === "self" || key === "parent" || key === "top" || key === "frames" || key === "document") continue;
        var child;
        try {
          child = value[key];
        } catch (_) {
          continue;
        }
        if (!looksTraversable(child)) continue;
        queue.push({
          value: child,
          path: item.path + "." + key,
          depth: item.depth + 1
        });
      }
    }
    log("stanza.scan", {
      visited: visited,
      wrapped: wrapped
    });
  }

  function hexHeadFromBytes(bytes, maxLen) {
    var n = Math.min(bytes.length || 0, maxLen || 32);
    var out = "";
    for (var i = 0; i < n; i++) {
      var h = bytes[i].toString(16);
      if (h.length < 2) h = "0" + h;
      out += h;
    }
    return out;
  }

  function hexTailFromBytes(bytes, maxLen) {
    var total = bytes.length || 0;
    var n = Math.min(total, maxLen || 32);
    var start = total - n;
    if (start < 0) start = 0;
    var out = "";
    for (var i = start; i < total; i++) {
      var h = bytes[i].toString(16);
      if (h.length < 2) h = "0" + h;
      out += h;
    }
    return out;
  }

  function describeBinary(data) {
    try {
      if (data instanceof ArrayBuffer) {
        var u8 = new Uint8Array(data);
        var lenHdr = null;
        var lenHdrMatches = null;
        var prefix3Hex = "";
        if (u8.length >= 3) {
          prefix3Hex = hexHeadFromBytes(u8, 3);
          lenHdr = ((u8[0] << 16) | (u8[1] << 8) | u8[2]) >>> 0;
          lenHdrMatches = lenHdr === (u8.length - 3);
        }
        var info = {
          type: "arraybuffer",
          len: data.byteLength,
          head_hex: hexHeadFromBytes(u8, 48),
          tail_hex: hexTailFromBytes(u8, 48),
          head_ascii: asciiHeadFromBytes(u8, 16),
          prefix3_hex: prefix3Hex,
          len_hdr: lenHdr,
          len_hdr_matches: lenHdrMatches
        };
        if (u8.length <= 256) {
          info.full_hex = hexHeadFromBytes(u8, u8.length);
        }
        return info;
      }
      if (globalThis.Blob && data instanceof Blob) {
        return {
          type: "blob",
          len: data.size
        };
      }
    } catch (_) {}
    return null;
  }

  function describeAny(data) {
    try {
      if (typeof data === "string") {
        return {
          type: "string",
          len: data.length,
          head: data.slice(0, 240)
        };
      }
      var bin = describeBinary(data);
      if (bin) return bin;
      if (Array.isArray(data)) {
        return {
          type: "array",
          len: data.length
        };
      }
      if (data && typeof data === "object") {
        return {
          type: Object.prototype.toString.call(data),
          keys: Object.keys(data).slice(0, 12),
          json: safe(data).slice(0, 240)
        };
      }
      return {
        type: typeof data,
        value: String(data)
      };
    } catch (_) {
      return {
        type: typeof data
      };
    }
  }

  function shouldLogMessagePayload(info) {
    if (!info) return false;
    if (info.type === "object" && info.value === "null") return false;
    return true;
  }

  function wrapAsync(proto, name) {
    if (!proto || typeof proto[name] !== "function") {
      return;
    }
    const original = proto[name];
    if (original[WRAP_TAG]) {
      return;
    }
    const wrapped = function (...args) {
      log(name + ":call", args);
      try {
        const result = original.apply(this, args);
        if (result && typeof result.then === "function") {
          return result.then(
            (value) => {
              log(name + ":ok", { value: value || null });
              return value;
            },
            (err) => {
              log(name + ":err", { message: err && err.message ? err.message : String(err) });
              throw err;
            }
          );
        }
        log(name + ":ok", { value: result || null });
        return result;
      } catch (err) {
        log(name + ":err", { message: err && err.message ? err.message : String(err) });
        throw err;
      }
    };
    wrapped[WRAP_TAG] = true;
    proto[name] = wrapped;
  }

  function describeSDP(desc) {
    if (!desc) return null;
    return {
      type: desc.type || null,
      sdp_len: desc.sdp ? desc.sdp.length : 0,
      sdp_head: desc.sdp ? desc.sdp.slice(0, 240) : ""
    };
  }

  function describeCandidate(ev) {
    const c = ev && ev.candidate ? ev.candidate : null;
    if (!c) return { candidate: null };
    return {
      candidate: c.candidate || null,
      sdpMid: c.sdpMid || null,
      sdpMLineIndex: c.sdpMLineIndex
    };
  }

  const NativePC = globalThis.RTCPeerConnection || globalThis.webkitRTCPeerConnection;
  if (typeof NativePC === "function") {
    const WrappedPC = function (...args) {
      const pc = new NativePC(...args);
      try {
        log("RTCPeerConnection:new", { config: args[0] || null });
      } catch (_) {}

      pc.addEventListener("icecandidate", (ev) => log("icecandidate", describeCandidate(ev)));
      pc.addEventListener("icecandidateerror", (ev) =>
        log("icecandidateerror", {
          url: ev && ev.url || null,
          hostCandidate: ev && ev.hostCandidate || null,
          address: ev && ev.address || null,
          port: ev && ev.port || null,
          errorCode: ev && ev.errorCode || null,
          errorText: ev && ev.errorText || null
        })
      );
      pc.addEventListener("iceconnectionstatechange", () =>
        log("iceconnectionstatechange", { state: pc.iceConnectionState })
      );
      pc.addEventListener("connectionstatechange", () =>
        log("connectionstatechange", { state: pc.connectionState })
      );
      pc.addEventListener("signalingstatechange", () =>
        log("signalingstatechange", { state: pc.signalingState })
      );
      pc.addEventListener("track", (ev) =>
        log("track", {
          kind: ev && ev.track ? ev.track.kind : null,
          id: ev && ev.track ? ev.track.id : null
        })
      );
      pc.addEventListener("datachannel", (ev) =>
        log("datachannel", {
          label: ev && ev.channel ? ev.channel.label : null,
          id: ev && ev.channel ? ev.channel.id : null
        })
      );

      return pc;
    };

    WrappedPC.prototype = NativePC.prototype;
    Object.setPrototypeOf(WrappedPC, NativePC);
    globalThis.RTCPeerConnection = WrappedPC;
    if (globalThis.webkitRTCPeerConnection) {
      globalThis.webkitRTCPeerConnection = WrappedPC;
    }

    wrapAsync(NativePC.prototype, "createOffer");
    wrapAsync(NativePC.prototype, "createAnswer");

    const origSetLocal = NativePC.prototype.setLocalDescription;
    if (typeof origSetLocal === "function" && !origSetLocal[WRAP_TAG]) {
      const wrapped = function (...args) {
        log("setLocalDescription:call", describeSDP(args[0]));
        return origSetLocal.apply(this, args);
      };
      wrapped[WRAP_TAG] = true;
      NativePC.prototype.setLocalDescription = wrapped;
    }

    const origSetRemote = NativePC.prototype.setRemoteDescription;
    if (typeof origSetRemote === "function" && !origSetRemote[WRAP_TAG]) {
      const wrapped = function (...args) {
        log("setRemoteDescription:call", describeSDP(args[0]));
        return origSetRemote.apply(this, args);
      };
      wrapped[WRAP_TAG] = true;
      NativePC.prototype.setRemoteDescription = wrapped;
    }

    const origAddIce = NativePC.prototype.addIceCandidate;
    if (typeof origAddIce === "function" && !origAddIce[WRAP_TAG]) {
      const wrapped = function (...args) {
        log("addIceCandidate:call", args[0] || null);
        return origAddIce.apply(this, args);
      };
      wrapped[WRAP_TAG] = true;
      NativePC.prototype.addIceCandidate = wrapped;
    }
  }

  if (globalThis.WebSocket && globalThis.WebSocket.prototype && typeof globalThis.WebSocket.prototype.send === "function") {
    const NativeWS = globalThis.WebSocket;
    const wsProto = globalThis.WebSocket.prototype;
    const origSend = globalThis.WebSocket.prototype.send;
    if (!origSend[WRAP_TAG]) {
      const wrapped = function (data) {
        let info = {};
        try {
          if (typeof data === "string") {
            info = { type: "string", len: data.length, head: data.slice(0, 240) };
          } else {
            info = describeBinary(data) || { type: typeof data };
          }
        } catch (_) {}
        log("WebSocket.send", info);
        return origSend.apply(this, arguments);
      };
      wrapped[WRAP_TAG] = true;
      wsProto.send = wrapped;
    }

    const origDispatchEvent = wsProto.dispatchEvent;
    if (typeof origDispatchEvent === "function" && !origDispatchEvent[WRAP_TAG]) {
      const wrappedDispatchEvent = function (ev) {
        try {
          if (ev && ev.type === "message") {
            let info = {};
            try {
              if (typeof ev.data === "string") {
                info = { type: "string", len: ev.data.length, head: ev.data.slice(0, 240) };
              } else {
                info = describeBinary(ev.data) || { type: typeof ev.data };
              }
            } catch (_) {}
            log("WebSocket.message", info);
          } else if (ev && ev.type === "open") {
            log("WebSocket.open", {});
          } else if (ev && ev.type === "close") {
            log("WebSocket.close", {
              code: ev.code || null,
              reason: ev.reason || null,
              wasClean: typeof ev.wasClean === "boolean" ? ev.wasClean : null
            });
          } else if (ev && ev.type === "error") {
            log("WebSocket.error", {});
          }
        } catch (_) {}
        return origDispatchEvent.apply(this, arguments);
      };
      wrappedDispatchEvent[WRAP_TAG] = true;
      wsProto.dispatchEvent = wrappedDispatchEvent;
    }

    const NativeEventTarget = globalThis.EventTarget;
    if (NativeEventTarget && NativeEventTarget.prototype && typeof NativeEventTarget.prototype.dispatchEvent === "function") {
      const origEventTargetDispatch = NativeEventTarget.prototype.dispatchEvent;
      if (!origEventTargetDispatch[WRAP_TAG]) {
        const wrappedEventTargetDispatch = function (ev) {
          try {
            if (this instanceof NativeWS) {
              if (ev && ev.type === "message") {
                let info = {};
                try {
                  if (typeof ev.data === "string") {
                    info = { type: "string", len: ev.data.length, head: ev.data.slice(0, 240) };
                  } else {
                    info = describeBinary(ev.data) || { type: typeof ev.data };
                  }
                } catch (_) {}
                log("WebSocket.message", info);
              } else if (ev && ev.type === "open") {
                log("WebSocket.open", {});
              } else if (ev && ev.type === "close") {
                log("WebSocket.close", {
                  code: ev.code || null,
                  reason: ev.reason || null,
                  wasClean: typeof ev.wasClean === "boolean" ? ev.wasClean : null
                });
              } else if (ev && ev.type === "error") {
                log("WebSocket.error", {});
              }
            }
          } catch (_) {}
          return origEventTargetDispatch.apply(this, arguments);
        };
        wrappedEventTargetDispatch[WRAP_TAG] = true;
        NativeEventTarget.prototype.dispatchEvent = wrappedEventTargetDispatch;
      }
    }

    if (!NativeWS[WRAP_CTOR_TAG]) {
      const WrappedWS = function (...args) {
        const ws = new NativeWS(...args);
        try {
          log("WebSocket:new", { url: args[0] || null });
        } catch (_) {}
        ws.addEventListener("message", function (ev) {
          let info = {};
          try {
            if (typeof ev.data === "string") {
              info = { type: "string", len: ev.data.length, head: ev.data.slice(0, 240) };
            } else {
              info = describeBinary(ev.data) || { type: typeof ev.data };
            }
          } catch (_) {}
          log("WebSocket.message", info);
        });
        return ws;
      };
      WrappedWS.prototype = NativeWS.prototype;
      Object.setPrototypeOf(WrappedWS, NativeWS);
      WrappedWS[WRAP_CTOR_TAG] = true;
      globalThis.WebSocket = WrappedWS;
    }
  }

  function wrapMessageEndpoint(ctor, name) {
    if (!ctor || !ctor.prototype) return;
    var proto = ctor.prototype;
    if (typeof proto.postMessage === "function" && !proto.postMessage[WRAP_TAG]) {
      var origPost = proto.postMessage;
      var wrappedPost = function (data) {
        var info = describeAny(data);
        if (shouldLogMessagePayload(info)) {
          log(name + ".postMessage", info);
        }
        return origPost.apply(this, arguments);
      };
      wrappedPost[WRAP_TAG] = true;
      proto.postMessage = wrappedPost;
    }

    if (typeof proto.dispatchEvent === "function" && !proto.dispatchEvent[WRAP_TAG]) {
      var origDispatch = proto.dispatchEvent;
      var wrappedDispatch = function (ev) {
        try {
          if (ev && ev.type === "message") {
            var info = describeAny(ev.data);
            if (shouldLogMessagePayload(info)) {
              log(name + ".message", info);
            }
          } else if (ev && ev.type === "messageerror") {
            log(name + ".messageerror", {});
          }
        } catch (_) {}
        return origDispatch.apply(this, arguments);
      };
      wrappedDispatch[WRAP_TAG] = true;
      proto.dispatchEvent = wrappedDispatch;
    }
  }

  wrapMessageEndpoint(globalThis.MessagePort, "MessagePort");
  wrapMessageEndpoint(globalThis.BroadcastChannel, "BroadcastChannel");

  if (globalThis.Worker && globalThis.Worker.prototype) {
    var NativeWorker = globalThis.Worker;
    wrapMessageEndpoint(NativeWorker, "Worker");
    if (!NativeWorker[WRAP_CTOR_TAG]) {
      var WrappedWorker = function () {
        var worker = Reflect.construct(NativeWorker, arguments, WrappedWorker);
        try {
          log("Worker:new", { script: arguments[0] || null });
        } catch (_) {}
        try {
          worker.addEventListener("message", function (ev) {
            var info = describeAny(ev.data);
            if (shouldLogMessagePayload(info)) {
              log("Worker.message", info);
            }
          });
          worker.addEventListener("messageerror", function () {
            log("Worker.messageerror", {});
          });
        } catch (_) {}
        return worker;
      };
      WrappedWorker.prototype = NativeWorker.prototype;
      Object.setPrototypeOf(WrappedWorker, NativeWorker);
      WrappedWorker[WRAP_CTOR_TAG] = true;
      globalThis.Worker = WrappedWorker;
    }
  }

  if (globalThis.SharedWorker) {
    var NativeSharedWorker = globalThis.SharedWorker;
    if (!NativeSharedWorker[WRAP_CTOR_TAG]) {
      var WrappedSharedWorker = function () {
        var worker = Reflect.construct(NativeSharedWorker, arguments, WrappedSharedWorker);
        try {
          log("SharedWorker:new", { script: arguments[0] || null, name: arguments[1] || null });
        } catch (_) {}
        try {
          if (worker && worker.port) {
            log("SharedWorker:port", {});
            worker.port.addEventListener("message", function (ev) {
              log("SharedWorker.port.message", describeAny(ev.data));
            });
            if (typeof worker.port.start === "function") {
              worker.port.start();
            }
          }
        } catch (_) {}
        return worker;
      };
      WrappedSharedWorker.prototype = NativeSharedWorker.prototype;
      Object.setPrototypeOf(WrappedSharedWorker, NativeSharedWorker);
      WrappedSharedWorker[WRAP_CTOR_TAG] = true;
      globalThis.SharedWorker = WrappedSharedWorker;
    }
  }

  if (globalThis.chrome && globalThis.chrome.webview) {
    try {
      log("chrome.webview.present", {
        keys: Object.keys(globalThis.chrome.webview).slice(0, 20)
      });
    } catch (_) {}

    if (typeof globalThis.chrome.webview.postMessage === "function" && !globalThis.chrome.webview.postMessage[WRAP_TAG]) {
      var origWVPost = globalThis.chrome.webview.postMessage;
      var wrappedWVPost = function (data) {
        log("chrome.webview.postMessage", describeAny(data));
        return origWVPost.apply(this, arguments);
      };
      wrappedWVPost[WRAP_TAG] = true;
      globalThis.chrome.webview.postMessage = wrappedWVPost;
    }

    if (typeof globalThis.chrome.webview.addEventListener === "function" && !globalThis.chrome.webview.addEventListener[WRAP_TAG]) {
      var origWVAdd = globalThis.chrome.webview.addEventListener;
      var wrappedWVAdd = function (type, listener) {
        log("chrome.webview.addEventListener", { type: type });
        return origWVAdd.apply(this, arguments);
      };
      wrappedWVAdd[WRAP_TAG] = true;
      globalThis.chrome.webview.addEventListener = wrappedWVAdd;
    }
  }

  if (globalThis.TextDecoder && globalThis.TextDecoder.prototype && typeof globalThis.TextDecoder.prototype.decode === "function") {
    var tdProto = globalThis.TextDecoder.prototype;
    var origTDDecode = tdProto.decode;
    if (!origTDDecode[WRAP_TAG]) {
      var wrappedTDDecode = function () {
        var out = origTDDecode.apply(this, arguments);
        try {
          logCallText("TextDecoder.decode.calltext", out, {
            encoding: this && this.encoding ? this.encoding : null
          });
        } catch (_) {}
        return out;
      };
      wrappedTDDecode[WRAP_TAG] = true;
      tdProto.decode = wrappedTDDecode;
    }
  }

  if (globalThis.XMLSerializer && globalThis.XMLSerializer.prototype && typeof globalThis.XMLSerializer.prototype.serializeToString === "function") {
    var xsProto = globalThis.XMLSerializer.prototype;
    var origSerializeToString = xsProto.serializeToString;
    if (!origSerializeToString[WRAP_TAG]) {
      var wrappedSerializeToString = function () {
        var out = origSerializeToString.apply(this, arguments);
        try {
          logCallText("XMLSerializer.serializeToString.calltext", out, {});
        } catch (_) {}
        return out;
      };
      wrappedSerializeToString[WRAP_TAG] = true;
      xsProto.serializeToString = wrappedSerializeToString;
    }
  }

  if (globalThis.DOMParser && globalThis.DOMParser.prototype && typeof globalThis.DOMParser.prototype.parseFromString === "function") {
    var dpProto = globalThis.DOMParser.prototype;
    var origParseFromString = dpProto.parseFromString;
    if (!origParseFromString[WRAP_TAG]) {
      var wrappedParseFromString = function (source, mimeType) {
        try {
          logCallText("DOMParser.parseFromString.calltext", source, {
            mimeType: mimeType || null
          });
        } catch (_) {}
        return origParseFromString.apply(this, arguments);
      };
      wrappedParseFromString[WRAP_TAG] = true;
      dpProto.parseFromString = wrappedParseFromString;
    }
  }

  if (typeof globalThis.postMessage === "function" && !globalThis.postMessage[WRAP_TAG]) {
    var origWindowPostMessage = globalThis.postMessage;
    var wrappedWindowPostMessage = function (data, targetOrigin, transfer) {
      log("window.postMessage", {
        payload: describeAny(data),
        targetOrigin: targetOrigin || null
      });
      return origWindowPostMessage.apply(this, arguments);
    };
    wrappedWindowPostMessage[WRAP_TAG] = true;
    globalThis.postMessage = wrappedWindowPostMessage;
  }

  log("installed", {
    version: WA_MON_VERSION,
    hasRTCPeerConnection: typeof NativePC === "function",
    hasWebSocket: typeof globalThis.WebSocket === "function"
  });

  try {
    scanForStanzaHooks();
    setTimeout(scanForStanzaHooks, 1500);
    setTimeout(scanForStanzaHooks, 5000);
    setInterval(scanForStanzaHooks, 15000);
  } catch (err) {
    log("stanza.scan.err", {
      message: err && err.message ? err.message : String(err)
    });
  }
})();

(function(){
  try {
    if (globalThis.__WA_STACK_MON_V1) return;
    globalThis.__WA_STACK_MON_V1 = true;

    function waStackShort() {
      try {
        var s = (new Error()).stack || '';
        var lines = String(s).split(/\r?\n/).slice(2, 8).map(function(x){ return String(x).trim(); });
        return lines.join(' | ');
      } catch (e) {
        return 'stack_unavailable';
      }
    }

    function waShouldTraceLen(len) {
      return [37,41,42,45,46,47,48,51,59,60,65,69,70,102,108,138,168,195,203,219,408,430,659,724,900,1280,2572,2683,5223].indexOf(len) >= 0;
    }

    var wsProto = globalThis.WebSocket && globalThis.WebSocket.prototype;
    if (wsProto && !wsProto.__wa_stack_send_wrapped) {
      var origSend = wsProto.send;
      Object.defineProperty(wsProto, '__wa_stack_send_wrapped', { value: true, configurable: true });
      wsProto.send = function(data) {
        try {
          var len = 0;
          if (data instanceof ArrayBuffer) len = data.byteLength;
          else if (ArrayBuffer.isView && ArrayBuffer.isView(data)) len = data.byteLength || data.length || 0;
          else if (data && typeof data.size === 'number') len = data.size;
          if (waShouldTraceLen(len)) {
            console.log('[WA-MON]', 'WebSocket.send.stack', JSON.stringify({ len: len, stack: waStackShort() }));
          }
        } catch (e) {}
        return origSend.apply(this, arguments);
      };
    }

    var workerProto = globalThis.Worker && globalThis.Worker.prototype;
    if (workerProto && !workerProto.__wa_stack_post_wrapped) {
      var origPost = workerProto.postMessage;
      Object.defineProperty(workerProto, '__wa_stack_post_wrapped', { value: true, configurable: true });
      workerProto.postMessage = function(msg) {
        try {
          var len = 0;
          if (msg instanceof ArrayBuffer) len = msg.byteLength;
          else if (ArrayBuffer.isView && ArrayBuffer.isView(msg)) len = msg.byteLength || msg.length || 0;
          if (len && waShouldTraceLen(len)) {
            console.log('[WA-MON]', 'Worker.postMessage.stack', JSON.stringify({ len: len, stack: waStackShort() }));
          }
        } catch (e) {}
        return origPost.apply(this, arguments);
      };
    }

    var mpProto = globalThis.MessagePort && globalThis.MessagePort.prototype;
    if (mpProto && !mpProto.__wa_stack_post_wrapped) {
      var origMPPost = mpProto.postMessage;
      Object.defineProperty(mpProto, '__wa_stack_post_wrapped', { value: true, configurable: true });
      mpProto.postMessage = function(msg) {
        try {
          var len = 0;
          if (msg instanceof ArrayBuffer) len = msg.byteLength;
          else if (ArrayBuffer.isView && ArrayBuffer.isView(msg)) len = msg.byteLength || msg.length || 0;
          if (len && waShouldTraceLen(len)) {
            console.log('[WA-MON]', 'MessagePort.postMessage.stack', JSON.stringify({ len: len, stack: waStackShort() }));
          }
        } catch (e) {}
        return origMPPost.apply(this, arguments);
      };
    }
  } catch (e) {
    console.log('[WA-MON]', 'stackhook:error', String(e));
  }
})();(function(){
  if (globalThis.__WA_SUBTLE_PROTO_MON_V1) return;
  globalThis.__WA_SUBTLE_PROTO_MON_V1 = true;
  function j(v){ try { return JSON.stringify(v); } catch (_) { return '"<unserializable>"'; } }
  function bufInfo(v){
    try {
      if (v == null) return null;
      var u8;
      if (v instanceof ArrayBuffer) u8 = new Uint8Array(v);
      else if (ArrayBuffer.isView(v)) u8 = new Uint8Array(v.buffer, v.byteOffset, v.byteLength);
      else return { type: typeof v };
      var n = u8.byteLength;
      var lim = Math.min(n, 64);
      var hex = '';
      for (var i = 0; i < lim; i++) { var h = u8[i].toString(16); if (h.length < 2) h = '0' + h; hex += h; }
      return { len: n, head_hex: hex };
    } catch (e) {
      return { err: String(e) };
    }
  }
  function algInfo(a){
    try {
      if (!a) return null;
      if (typeof a === 'string') return { name: a };
      var o = {};
      Object.keys(a).forEach(function(k){
        var v = a[k];
        if (typeof v === 'string' || typeof v === 'number' || typeof v === 'boolean') o[k] = v;
        else if (v instanceof ArrayBuffer || ArrayBuffer.isView(v)) o[k] = bufInfo(v);
      });
      return o;
    } catch (e) {
      return { err: String(e) };
    }
  }
  function stackStr(){
    try {
      return String(new Error().stack || '').split('\n').slice(2, 8).join(' | ');
    } catch (_) {
      return '';
    }
  }
  var proto = globalThis.SubtleCrypto && globalThis.SubtleCrypto.prototype;
  if (!proto) {
    console.log('[WA-MON] crypto.proto.missing {}');
    return;
  }
  ['encrypt','decrypt','importKey'].forEach(function(name){
    var orig = proto[name];
    if (typeof orig !== 'function' || orig.__waWrappedProtoV1) return;
    function wrapped(){
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
      try { console.log('[WA-MON] crypto.proto.' + name + ' ' + j(payload)); } catch (_) {}
      return orig.apply(this, arguments);
    }
    wrapped.__waWrappedProtoV1 = true;
    proto[name] = wrapped;
  });
  try { console.log('[WA-MON] crypto.proto.installed ' + j({ ok: true })); } catch (_) {}
})();
;(function(){
  if (globalThis.__WA_CRYPTO_DIRECT_MON_V2) return;
  globalThis.__WA_CRYPTO_DIRECT_MON_V2 = true;
  function j(v){ try { return JSON.stringify(v); } catch (_) { return '"<unserializable>"'; } }
  function bufInfo(v){
    try {
      if (v == null) return null;
      var u8;
      if (v instanceof ArrayBuffer) u8 = new Uint8Array(v);
      else if (ArrayBuffer.isView(v)) u8 = new Uint8Array(v.buffer, v.byteOffset, v.byteLength);
      else return { type: typeof v };
      var n = u8.byteLength;
      var lim = Math.min(n, 64);
      var hex = '';
      for (var i = 0; i < lim; i++) { var h = u8[i].toString(16); if (h.length < 2) h = '0' + h; hex += h; }
      return { len: n, head_hex: hex };
    } catch (e) {
      return { err: String(e) };
    }
  }
  function algInfo(a){
    try {
      if (!a) return null;
      if (typeof a === 'string') return { name: a };
      var o = {};
      Object.keys(a).forEach(function(k){
        var v = a[k];
        if (typeof v === 'string' || typeof v === 'number' || typeof v === 'boolean') o[k] = v;
        else if (v instanceof ArrayBuffer || ArrayBuffer.isView(v)) o[k] = bufInfo(v);
      });
      return o;
    } catch (e) {
      return { err: String(e) };
    }
  }
  function stackStr(){
    try { return String(new Error().stack || '').split('\n').slice(2, 8).join(' | '); } catch (_) { return ''; }
  }
  var subtle = self.crypto && self.crypto.subtle;
  if (!subtle) {
    try { console.log('[WA-MON] crypto.direct.missing {}'); } catch (_) {}
    return;
  }
  ['encrypt','decrypt','importKey','sign'].forEach(function(name){
    var orig = subtle[name];
    if (typeof orig !== 'function' || orig.__waDirectWrappedV2) return;
    function wrapped(){
      var args = Array.prototype.slice.call(arguments);
      var payload = { method: name, stack: stackStr() };
      try {
        if (name === 'importKey') {
          payload.format = args[0];
          payload.keyData = bufInfo(args[1]);
          payload.algorithm = algInfo(args[2]);
          payload.extractable = !!args[3];
          payload.usages = args[4];
        } else if (name === 'sign') {
          payload.algorithm = algInfo(args[0]);
          payload.keyType = args[1] && args[1].type ? args[1].type : typeof args[1];
          payload.data = bufInfo(args[2]);
        } else {
          payload.algorithm = algInfo(args[0]);
          payload.keyType = args[1] && args[1].type ? args[1].type : typeof args[1];
          payload.data = bufInfo(args[2]);
        }
      } catch (e) {
        payload.err = String(e);
      }
      try { console.log('[WA-MON] crypto.direct.' + name + ' ' + j(payload)); } catch (_) {}
      return orig.apply(subtle, arguments);
    }
    wrapped.__waDirectWrappedV2 = true;
    subtle[name] = wrapped;
  });
  try { console.log('[WA-MON] crypto.direct.installed ' + j({ ok: true })); } catch (_) {}
})();
;(function(){
  if (globalThis.__WA_CRYPTO_RESULT_MON_V2) return;
  globalThis.__WA_CRYPTO_RESULT_MON_V2 = true;
  function j(v){ try { return JSON.stringify(v); } catch (_) { return '"<unserializable>"'; } }
  function bufInfo(v){
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
        for (var j = 0; j < n; j++) { var fh = u8[j].toString(16); if (fh.length < 2) fh = '0' + fh; full += fh; }
        info.full_hex = full;
      } else {
        var tail = '';
        var start = Math.max(0, n - 96);
        for (var k = start; k < n; k++) { var th = u8[k].toString(16); if (th.length < 2) th = '0' + th; tail += th; }
        info.tail_hex = tail;
      }
      return info;
    } catch (e) {
      return { err: String(e) };
    }
  }
  var subtle = self.crypto && self.crypto.subtle;
  if (!subtle) {
    try { console.log('[WA-MON] crypto.result.missing {}'); } catch (_) {}
    return;
  }
  ['encrypt','decrypt'].forEach(function(name){
    var orig = subtle[name];
    if (typeof orig !== 'function' || orig.__waResultWrappedV2) return;
    function wrapped(){
      var ret = orig.apply(subtle, arguments);
      try {
        if (ret && typeof ret.then === 'function') {
          ret.then(function(value){
            try { console.log('[WA-MON] crypto.result.' + name + ' ' + j(bufInfo(value))); } catch (_) {}
            return value;
          }, function(err){
            try { console.log('[WA-MON] crypto.result.' + name + '.error ' + j({ err: String(err) })); } catch (_) {}
          });
        }
      } catch (_) {}
      return ret;
    }
    wrapped.__waResultWrappedV2 = true;
    subtle[name] = wrapped;
  });
  try { console.log('[WA-MON] crypto.result.installed ' + j({ ok: true })); } catch (_) {}
})();
