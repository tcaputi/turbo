var Turbo = (function() {
    'use strict';

    var MSG_CMD_ON = 1,
        MSG_CMD_OFF = 2,
        MSG_CMD_SET = 3,
        MSG_CMD_UPDATE = 4,
        MSG_CMD_REMOVE = 5,
        MSG_CMD_TRANS_SET = 6,
        MSG_CMD_PUSH = 7,
        MSG_CMD_TRANS_GET = 8,
        MSG_CMD_AUTH = 9,
        MSG_CMD_UNAUTH = 10,
        MSG_CMD_ACK = 11;

    var EVENT_TYPE_VALUE = 0,
        EVENT_TYPE_CHILD_ADDED = 1,
        EVENT_TYPE_CHILD_CHANGED = 2,
        EVENT_TYPE_CHILD_MOVED = 3,
        EVENT_TYPE_CHILD_REMOVED = 4;

    var EVENT_TYPE_VALUE_STR = 'value',
        EVENT_TYPE_CHILD_ADDED_STR = 'child_added',
        EVENT_TYPE_CHILD_CHANGED_STR = 'child_changed',
        EVENT_TYPE_CHILD_MOVED_STR = 'child_moved',
        EVENT_TYPE_CHILD_REMOVED_STR = 'child_removed';

    var _ws = undefined;
    var _listeners = {};
    var _ack = 0;
    var _ackCallbacks = {};
    var _token = undefined;
    var _isOffline = true;
    var _offlineQueue = [];

    var _send = function _send(val) {
        console.log("Sending: ", val);
        if (_isOffline) {
            _offlineQueue.push(val);
        } else {
            if (!_ws) throw 'Must be connected to the server to send a message';
            _ws.send(val);
        }
    };

    var _connect = function _connect(url, onConnect, onError, onClose) {
        console.log('Connecting to ' + url);
        _ws = new WebSocket(url);

        _ws.onopen = function(evt) {
            _isOffline = false;
            if (onConnect) onConnect(evt);
            console.log('Connection opened.', evt);
        };
        _ws.onclose = function(evt) {
            _isOffline = true;
            if (onClose) onClose();
            console.log('Connection closed.', evt);
        };
        _ws.onerror = function(evt) {
            if (onError) onError(evt);
            console.log('Connection had an error.', evt);
        };
        _ws.onmessage = function(evt) {
            if (!evt) return;
            if (!evt.data) return;

            try {
                var msg = JSON.parse(evt.data);
                switch (msg.type) {
                    case MSG_CMD_ACK:
                        if (_ackCallbacks[msg.ack]) {
                            _ackCallbacks[msg.ack](msg.err, msg.res, msg.revision);
                            delete _ackCallbacks[msg.ack];
                        }
                        break;
                    default:
                        if (!msg.eventType || !msg.path) return; // Filter for 'on' events

                        if (_listeners[msg.path] && _listeners[msg.path][msg.eventType]) {
                            var listenerMap = _listeners[msg.path][msg.eventType];
                            for (var listenerRef in listenerMap) {
                                var listener;
                                if (listener = listenerMap[listenerRef]) {
                                    var context = listener.context || listenerRef;
                                    if (listener.callback) {
                                        listener.callback.call(context, new DataSnapShot(msg.value, this._url, msg.path));
                                    }
                                }
                            }
                        }
                }
            } catch (e) {
                console.log('Could not process message; json parsing is failing', e);
            }
        };
    };

    var _disconnect = function _disconnect() {
        console.log('Disconnecting');
        _isOffline = true;
        _ws.close();
        _ws = undefined;
    };

    var _attemptTransSet = function _attemptTransSet(path, value, rev, transform, done) {
        var ack = _ack++;
        _send(JSON.stringify({
            'cmd': MSG_CMD_TRANS_SET,
            'path': path,
            'revision': rev,
            'value': transform(value),
            'ack': ack
        }));
        _ackCallbacks[ack] = function(err, newValue, rev) {
            if (err === 'conflict') _attemptTransSet(path, newValue, rev, transform, done);
            else if (err) done(err);
            else done(undefined, newValue);
        };
    };

    var _eventType = function _eventType(eventTypeStr) {
        switch (eventTypeStr) {
            case EVENT_TYPE_VALUE_STR:
                return EVENT_TYPE_VALUE;
            case EVENT_TYPE_CHILD_ADDED_STR:
                return EVENT_TYPE_CHILD_ADDED;
            case EVENT_TYPE_CHILD_CHANGED_STR:
                return EVENT_TYPE_CHILD_CHANGED;
            case EVENT_TYPE_CHILD_MOVED_STR:
                return EVENT_TYPE_CHILD_MOVED;
            case EVENT_TYPE_CHILD_REMOVED:
                return EVENT_TYPE_CHILD_REMOVED;
        }
        return null;
    };

    var _sanitizeUrl = function _sanitizeUrl(url) {
        if (url.charAt(0) === '/') url = window.location.host + url;
        // Replace // with /
        url = url.replace(/(\w+):(\/+)/g, 'ws://');
        if (url.indexOf('ws://') !== 0) {
            url = 'ws://' + url;
        }
        // Cleans off trailing /
        if (url.charAt(url.length - 1) === '/' && url.length > 1) {
            return url.substring(0, url.length - 1);
        }
        return url;
    };

    var _sanitizePath = function _sanitizePath(path) {
        if (path === undefined || path === null) return '/';
        // Replace // with /
        path = path.replace(/\/{2,}/g, '/');
        // Make sure path starts with /
        if (path.charAt(0) !== '/') {
            path = '/' + path;
        }
        // Cleans off trailing /
        if (path.charAt(path.length - 1) === '/' && path.length > 1) {
            return path.substring(0, path.length - 1);
        }
        return path;
    };

    var _joinPaths = function _joinPaths(base, ext) {
        base = _sanitizePath(base);
        return base + '/' + ext;
    };

    var _flatten = function _flatten(basePath, obj, res) {
        if (obj === undefined || obj === null || !basePath || !res) return undefined;

        if (typeof obj === 'object') {
            for (var key in obj) {
                _flatten(basePath, obj[key], res);
            }
        } else if (Object.prototype.toString.call(obj) === '[object Array]') {
            for (var i = 0; i < obj.length; i++) {
                _flatten(_joinPaths(basePath, i), obj[i], res);
            }
        } else {
            res[basePath] = obj;
        }
    };

    var _inflate = function _inflate(basePath, obj) {
        if (obj === undefined || obj === null || !basePath) return undefined;
        var res = {},
            part, parts, currObj, i;

        for (var key in obj) {
            parts = _sanitizePath(path).split('/');
            currObj = res;
            i = 0;
            for (part in parts) {
                if (part !== '' && path !== undefined && path !== null) {
                    if (i < parts.length - 1) {
                        currObj[part] = (currObj = (currObj[part] || {}));
                    } else {
                        currObj[part] = obj[key];
                    }
                }
            }
        }
    };

    // Constructor
    var Client = function(url, path) {
        if (url === null || url === undefined || !(typeof url === 'string'))
            throw new Error('url was invalid');
        if (path === null || path === undefined || !(typeof path === 'string'))
            path = '/';

        this._url = _sanitizeUrl(url);
        this._path = _sanitizePath(path);

        if (!_ws) _connect(this._url, function(evt) {
            var msg;
            while ((msg = _offlineQueue.shift())) {
                _ws.send(msg);
            }
        });
    };

    Client.prototype.on = function(eventTypeStr, callback, cancelCallback, context) {
        var eventType = _eventType(eventTypeStr);
        if (!eventType && eventType !== 0)
            throw new Error('Unsupported event type \'' + eventTypeStr + '\'');

        if (!callback || typeof callback !== 'function') throw 'Callback was not a function';
        if (cancelCallback && !context && typeof cancelCallback === 'object') {
            context = cancelCallback;
            cancelCallback = undefined;
        }

        var self = this;
        var path = self._path;
        _send(JSON.stringify({
            'cmd': MSG_CMD_ON,
            'eventType': eventType,
            'path': path
        }));

        if (!_listeners[path]) _listeners[path] = {};
        if (!_listeners[path][eventType]) _listeners[path][eventType] = {};
        _listeners[path][eventType][self] = {
            callback: callback,
            context: context,
            cancelCallback: cancelCallback
        };

        return callback;
    };

    Client.prototype.off = function(eventType, callback, context) {
        var eventType = _eventType(eventTypeStr);
        if (!eventType && eventType !== 0)
            throw 'Unsupported event type \'' + eventTypeStr + '\'';

        if (!callback || typeof callback !== 'function') throw 'Callback was not a function';

        var self = this;
        var path = self._path;

        if (!_listeners[path]) return;
        if (!_listeners[path][eventType]) return;
        if (!_listeners[path][eventType][self]) return;
        delete _listeners[path][eventType][self];

        _send(JSON.stringify({
            'cmd': MSG_CMD_OFF,
            'eventType': eventType,
            'path': path
        }));
    };

    Client.prototype.child = function(childPath) {
        if (!childPath) return this;
        return new Client(this._url, _joinPaths(this._path, childPath));
    };

    Client.prototype.parent = function() {
        if (this._path === '/') return this;
        var newPath = this._path.slice(0, this._path.lastIndexOf('/'));
        return new Client(this._url, newPath);
    };

    Client.prototype.root = function() {
        if (this._path === '/') return this;
        return new Client(this._url, '/');
    };

    Client.prototype.toString = function() {
        return this._url + this._path;
    };

    Client.prototype.set = function(value, onComplete) {
        var self = this;
        var ack = _ack++;
        var deltas = {};

        _flatten(this._path, value, deltas);
        _send(JSON.stringify({
            'cmd': MSG_CMD_SET,
            'path': self._path,
            'deltas': deltas,
            'ack': ack
        }));
        _ackCallbacks[ack] = onComplete;
    };

    Client.prototype.update = function(value, onComplete) {
        var self = this;
        var ack = _ack++;
        var deltas = {};

        _flatten(this._path, value, deltas);
        _send(JSON.stringify({
            'cmd': MSG_CMD_UPDATE,
            'path': self._path,
            'deltas': deltas,
            'ack': ack
        }));
        _ackCallbacks[ack] = onComplete;
    };

    Client.prototype.name = function() {
        return this._path.split('/').pop();
    };

    Client.prototype.setWithPriority = function(newVal, newPriority, onComplete) {
        throw 'Turbo does not support setWithPriority(...) right now';
    };

    Client.prototype.remove = function(onComplete) {
        var self = this;
        var ack = _ack++;
        _send(JSON.stringify({
            'cmd': MSG_CMD_REMOVE,
            'path': self._path,
            'ack': ack
        }));
        _ackCallbacks[ack] = onComplete;
    };

    Client.prototype.transaction = function(transactionUpdate, onComplete, applyLocally) {
        var self = this;
        var ack = _ack++;
        _send(JSON.stringify({
            'cmd': MSG_CMD_TRANS_GET,
            'path': self._path,
            'ack': ack
        }));
        _ackCallbacks[ack] = function(err, value, rev) {
            if (err) onComplete(err);
            else _attemptTransSet(self._path, value, rev, transactionUpdate, onComplete);
        };
    };

    Client.prototype.setPriority = function(priority, opt_onComplete) {
        throw 'Turbo does not support setPriority(...) right now';
    };

    Client.prototype.push = function(value, onComplete) {
        var self = this;
        var ack = _ack++;
        _send(JSON.stringify({
            'cmd': MSG_CMD_PUSH,
            'path': self._path,
            'value': value,
            'ack': ack
        }));
        _ackCallbacks[ack] = onComplete;
    };

    Client.prototype.onDisconnect = function() {
        throw 'Turbo does not support onDisconnect(...) right now';
    };

    Client.prototype.removeOnDisconnect = function() {
        throw 'Turbo does not support removeOnDisconnect(...) right now';
    };

    Client.prototype.setOnDisconnect = function(onDc) {
        throw 'Turbo does not support setOnDisconnect(...) right now';
    };

    Client.prototype.auth = function(cred, onComplete, onCancel) {
        var ack = _ack++;
        _send(JSON.stringify({
            'cmd': MSG_CMD_AUTH,
            'cred': cred,
            'ack': ack
        }));
        _ackCallbacks[ack] = onComplete;
    };

    Client.prototype.unauth = function(onComplete) {
        if (!_token) throw 'Cannot unauth if not authed yet';

        var ack = _ack++;
        _send(JSON.stringify({
            'cmd': MSG_CMD_UNAUTH,
            'token': _token,
            'ack': ack
        }));
        _ackCallbacks[ack] = function(err, res) {
            _token = undefined;
            onComplete(err, res);
        };
    };

    Client.prototype.goOffline = function() {
        _disconnect();
    };

    Client.prototype.goOnline = function() {
        _connect(this._url, function(evt) {
            var msg;
            while ((msg = _offlineQueue.shift())) {
                _ws.send(msg);
            }
        });
    };

    Client.prototype.enableLogging = function(logger, persistent) {
        throw 'Turbo does not support enableLogging(...) right now';
    };

    function DataSnapshot(baseObj, url, path) {
        this._baseObj = baseObj;
        this._url = url;
        this._path = path;
    }

    DataSnapshot.prototype.val = function() {
        return this._baseObj;
    };

    DataSnapshot.prototype.child = function(childName) {
        return new DataSnapshot(this._baseObj[childName]);
    }

    DataSnapshot.prototype.forEach = function(childAction) {
        for (var child in this._baseObj) {
            childAction(child);
        }
    };

    DataSnapshot.prototype.hasChild = function(childName) {
        return !!this._baseObj[childName];
    };

    DataSnapshot.prototype.hasChildren = function() {
        return Object.keys(this._baseObj).length != 0;
    };

    DataSnapshot.prototype.name = function() {
        return this._baseObj.split('/').pop();
    };

    DataSnapshot.prototype.numChildren = function() {
        return Object.keys(this._baseObj).length;
    };

    DataSnapshot.prototype.ref = function() {
        return new Client(this._url, this._path);
    };

    DataSnapshot.prototype.getPriority = function() {
        //TODO: are we doing this?
    };

    DataSnapshot.prototype.exportVal = function() {
        //TODO: are we doing priority? if not this is the same as val()
    };

    return Client;
})();