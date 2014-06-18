var Turbo = (function () {
    'use strict';

    var CMD_ON          = 'on';
    var CMD_SET         = 'set';
    var CMD_UPDATE      = 'update';
    var CMD_REMOVE      = 'remove';
    var CMD_TRANS_SET   = 'transSet';
    var CMD_PUSH        = 'push';
    var CMD_GET         = 'get';
    var CMD_AUTH        = 'auth';
    var CMD_UNAUTH      = 'unauth';

    var MSG_TYPE_ACK        = 'ack';
    var MSG_TYPE_AUTH_ACK   = 'authAck';

    var _ws = undefined;
    var _listeners = [];
    var _ack = 0;
    var _ackCallbacks = {};
    var _token = undefined;
    var _isOffline = true;
    var _offlineQueue = [];

    var _send = function _send(val) {
        if (_isOffline) {
            _offlineQueue.push(val);
        } else {
            if (!_ws) throw 'Must be connected to the server to send a message';
            _ws.send(val);
        }
    };

    var _connect = function _connect(url, onConnect) {
        console.log('Connecting to ' + url);
        _ws = new WebSocket(url);
        _ws.onopen = function(evt) {
            if (onConnect) onConnect(evt);
            _isOffline = false;
            console.log('Connection opened.', evt);
        };
        _ws.onclose = function(evt) {
            _isOffline = true;
            console.log('Connection closed.', evt);
        };
        _ws.onerror = function(evt) {
            console.log('Connection had an error.', evt);
        };
        _ws.onmessage = function(evt) {
            try {
                var msg = JSON.parse(evt);
                switch (msg.type) {
                    case MSG_TYPE_ACK:
                        if (_ackCallbacks[msg.ack]) {
                            _ackCallbacks[msg.ack](msg.err, msg.res);
                            delete _ackCallbacks[msg.ack];
                        }
                        break;
                    case MSG_TYPE_AUTH_ACK:
                        if (_ackCallbacks[msg.ack]) {
                            _token = msg.token;
                            _ackCallbacks[msg.ack](msg.err, msg.res);
                            delete _ackCallbacks[msg.ack];
                        }
                        break;
                }
            } catch (e) {
                console.log('Could not process message; json parsing is failing')
            }
        };
    };

    var _disconnect = function _disconnect() {
        console.log('Disconnecting');
        _isOffline = true;
        _ws.close();
        _ws = undefined;
    };

    var _attemptTransSet = function _attemptTransSet(path, value, transform, done) {
        ack = _ack++;
        _send(JSON.stringify({
            'cmd': CMD_TRANS_SET,
            'path': path,
            'basis': value,
            'value': transform(value),
            'ack': ack
        }));
        _ackCallbacks[ack] = function(err, newValue) {
            if (err === 'conflict') _attemptTransSet(path, newValue, transform, done);
            else done(undefined, newValue);
        };
    };

    var Client = function(url, path) {
        if (url === null || url === undefined || !(typeof url === 'string'))
            url = window.location.host;
        if (path === null || path === undefined || !(typeof path === 'string'))
            path = '/';

        if (url.charAt(0) === '/') url = window.location.host + url;
        if (url.indexOf('ws://') !== 0) url = 'ws://' + url;
        if (url.charAt(url.length - 1) === '/') url = url.slice(0, -1);

        this._url = url;
        this._path = path;

        if (!_ws) _connect(url);
    };

    Client.prototype.on = function(eventType, callback, cancelCallback, context) {
        if (eventType !== 'value' &&
            eventType !== 'child_added' &&
            eventType !== 'child_changed' &&
            eventType !== 'child_removed' &&
            eventType !== 'child_moved')
            throw 'Unsupported event type \'' + eventType + '\'';
        if (!callback || typeof callback !== 'function') throw 'Callback was not a function';
        if (cancelCallback && !context && typeof cancelCallback === 'object') {
            context = cancelCallback;
            cancelCallback = undefined;
        }

        var self = this;
        _send(JSON.stringify({
            'cmd': CMD_ON,
            'eventType': eventType,
            'path': self._path
        }));

        if (!_listeners[this._path]) _listeners[this._path] = {};
        if (!_listeners[this._path][eventType]) _listeners[this._path][eventType] = [];
        _listeners[this._path][eventType].push({
            callback: callback,
            context: context,
            cancelCallback: cancelCallback
        });
    };

    Client.prototype.child = function(childPath) {
        var newPath = this._path === '/' ? ('/' + childPath) : (this._path + '/' + childPath);
        return new Client(this._url, newPath);
    };

    Client.prototype.parent = function() {
        if (this._path === '/') return this;
        var newPath = this._path.slice(0, this._path.lastIndexOf('/'));
        return new Client(this._url, newPath);
    };

    Client.prototype.root = function() {
        return new Client(this._url, '/');
    };

    Client.prototype.toString = function() {
        return this._path;
    };

    Client.prototype.set = function(value, onComplete) {
        var self = this;
        var ack = _ack++;
        _send(JSON.stringify({
            'cmd': CMD_SET,
            'path': self._path,
            'value': value,
            'ack': ack
        }));
        _ackCallbacks[ack] = onComplete;
    };

    Client.prototype.update = function(objectToMerge, onComplete) {
        var self = this;
        var ack = _ack++;
        _send(JSON.stringify({
            'cmd': CMD_UPDATE,
            'path': self._path,
            'value': objectToMerge,
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
            'cmd': CMD_REMOVE,
            'path': self._path,
            'ack': ack
        }));
        _ackCallbacks[ack] = onComplete;
    };

    Client.prototype.transaction = function(transactionUpdate, onComplete, applyLocally) {
        var self = this;
        var ack = _ack++;
        _send(JSON.stringify({
            'cmd': CMD_GET,
            'path': self._path,
            'ack': ack
        }));
        _ackCallbacks[ack] = function(err, value) {
            if (err) onComplete(err);
            else _attemptTransSet(self._path, value, onComplete);
        };
    };

    Client.prototype.setPriority = function(priority, opt_onComplete) {
        throw 'Turbo does not support setPriority(...) right now';
    };

    Client.prototype.push = function(value, onComplete) {
        var self = this;
        var ack = _ack++;
        _send(JSON.stringify({
            'cmd': CMD_PUSH,
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
            'cmd': CMD_AUTH,
            'cred': cred,
            'ack': ack
        }));
        _ackCallbacks[ack] = onComplete;
    };

    Client.prototype.unauth = function(onComplete) {
        if (!_token) throw 'Cannout unauth if not authed yet';

        var ack = _ack++;
        _send(JSON.stringify({
            'cmd': CMD_UNAUTH,
            'token': _token,
            'ack': ack
        }));
        _ackCallbacks[ack] = function(err, res) {
            _token = undefined;
            onComplete(err, res);
        };
    };

    Client.goOffline = function() {
        _disconnect();
    };

    Client.goOnline = function() {
        _connect(function (evt) {
            var msg;
            while ((msg = _offlineQueue.shift())) {
                _ws.send(JSON.parse(msg));
            }
        });
    };

    Client.enableLogging = function(logger, persistent) {
        throw 'Turbo does not support enableLogging(...) right now';
    };

    return Client;
})();