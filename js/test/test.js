var assert = chai.assert;
var testVal = {
    a: 'hi',
    b: 'there'
};

describe('Turbo', function() {
    describe('constructor', function() {
        it('should yell at us if we give it a non-string url', function() {
            assert.throw(function() {
                new Turbo();
            });
            assert.throw(function() {
                new Turbo({});
            });
        });

        it('should make a Turbo instance', function() {
            assert.doesNotThrow(function() {
                new Turbo('dsf94ouiyoiwruy.com', '/fake/path')
            });
        });
    });

    describe('#child', function() {
        it('should be able to change paths to a child element', function() {
            var turbo = new Turbo('http://127.0.0.1', '/test');
            var subTurbo = turbo.child('a');
            assert.equal(subTurbo._path, '/test/a');
        });

        it('should be able to change paths to a deep child element', function() {
            var turbo = new Turbo('http://127.0.0.1', '/test');
            var subTurbo = turbo.child('/a/b/c');
            assert.equal(subTurbo._path, '/test/a/b/c');
        });
    });

    describe('#parent', function() {
        it('should be able to change paths to a parent element', function() {
            var turbo = new Turbo('http://127.0.0.1', '/test');
            var subTurbo = turbo.child('a');
            var parentSubTurbo = subTurbo.parent();
            assert.equal(turbo._path, parentSubTurbo._path);
        });
    });

    describe('#root', function() {
        it('should be able to change paths to the root element', function() {
            var turbo = new Turbo('http://127.0.0.1', '/test');
            var subTurbo = turbo.child('a/b/c');
            var rootTurbo = subTurbo.root();
            assert.equal(rootTurbo._path, '/');
        });
    });

    describe('#name', function() {
        it('should be able to return the name of the path', function() {
            var turbo = new Turbo('http://127.0.0.1', '/test');
            var subPath = turbo.child('a/b/c');
            assert.equal(subPath.name(), 'c');
        });
    });

    describe('#toString', function() {
        it('should be able to return the url and path', function() {
            var turbo = new Turbo('http://127.0.0.1', '/test');
            assert.equal(turbo.toString(), 'ws://127.0.0.1/test');
        });
    });

    describe('#set and #on', function() {
        it('should be able to perform a set on the root object and tell everyone about it', function(done) {
            var turbo = new Turbo('http://127.0.0.1', '/test');
            turbo.on('value', function(snapshot) {
                assert.deepEqual(snapshot, testVal);
                assert.equal(snapshot._path, '/');
                done();
            });
            turbo.set(testVal);
        });

        it('should be able to add a new direct child to the root and tell the root listener', function(done) {
            var turbo = new Turbo('http://127.0.0.1', '/test');
            turbo.on('child_added', function(snapshot) {
                assert.deepEqual(snapshot.val(), testVal);
                assert.equal(snapshot._path, '/a');
                done();
            });
            turbo.child('a').set(testVal);
        });

        it('should be able to add a new deep child to the root and tell the root listener', function(done) {
            var turbo = new Turbo('http://127.0.0.1', '/test');
            turbo.on('child_added', function(snapshot) {
                assert.deepEqual(snapshot.val(), testVal);
                assert.equal(snapshot._path, '/a/b');
                done();
            });
            turbo.child('a').child('b').set(testVal);
        });

        it('should be able to alter a direct child to the root and tell the root listener', function(done) {
            var turbo = new Turbo('http://127.0.0.1', '/test');
            turbo.set(testVal);
            turbo.on('child_changed', function(snapshot) {
                assert.deepEqual(snapshot.val(), testVal);
                assert.equal(snapshot._path, '/a');
                done();
            });
            turbo.child('a').set(testVal);
        });

        //TODO: add child_removed listener tests for all cases it should be called
        it('should be able to remove a direct child to the root via set(null) and tell the root listener', function(done) {
            var turbo = new Turbo('http://127.0.0.1', '/test');
            turbo.set(testVal);
            turbo.on('child_removed', function(oldSnapshot) {
                assert.deepEqual(oldSnapshot.val(), testVal);
                assert.equal(snapshot._path, '/a');
                done();
            });
            turbo.child('a').set(null);
        });

        it('should be able to alter a direct child to the root and tell the root listener', function(done) {
            var turbo = new Turbo('http://127.0.0.1', '/test');
            turbo.set(testVal);
            turbo.on('child_changed', function(snapshot) {
                assert.deepEqual(snapshot.val(), {
                    a: 'new hi',
                    b: 'there'
                });
                assert.equal(snapshot._path, '/a');
                done();
            });
            turbo.child('a').set('new hi');
        });
    });

    describe('#off', function() {
        it('should be able to turn off a listener', function() {
            var turbo = new Turbo('http://127.0.0.1', '/test');
            var onValue = function(snapshot) {
                assert.fail('', '', 'this should not have been called, callback should be unregistered');
            };
            turbo.on('value', onValue);
            turbo.off('value', onValue);
            turbo.set(testVal);
        });
    });

    describe('#update', function() {
        it('should be able to perform an update on the root object, leaving existing data untocuhed, and tell everyone about it', function(done) {
            var turbo = new Turbo('http://127.0.0.1', '/test');
            turbo.set(testVal);
            turbo.on('value', function(snapshot) {
                var newTestVal = testVal;
                newTestVal.c = 'bye';
                assert.deepEqual(snapshot, newTestVal);
                assert.equal(snapshot._path, '/');
                done();
            });
            turbo.update({
                c: 'bye'
            });

        });
    });

    describe('#remove', function() {
        it('should be able to remove a direct child to the root and tell the root listener', function(done) {
            var turbo = new Turbo('http://127.0.0.1', '/test');
            turbo.set(testVal);
            turbo.on('child_removed', function(oldSnapshot) {
                assert.deepEqual(oldSnapshot.val(), testVal);
                assert.equal(snapshot._path, '/a');
                done();
            });
            turbo.child('a').remove();
        });
    });

    describe('#transaction', function() {});

    describe('#push', function() {});

    describe('#onDisconnect', function() {});

    describe('#removeOnDisconnect', function() {});

    describe('#setOnDisconnect', function() {});

    describe('#goOffline', function() {});

    describe('#goOnline', function() {});

    describe('#enableLogging', function() {});

    describe('#auth', function() {});

    describe('#unauth', function() {});

    describe('#setPriority', function() {});

    describe('#setWithPriority', function() {});
});