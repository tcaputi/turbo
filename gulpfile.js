var gulp = require('gulp');
var exec = require('child_process').exec;
var curr;

gulp.task('compile', function(cb) {
    exec('go build', function(err, stdout, stderr) {
        cb(err);
    });
});

gulp.task('run', ['compile'], function(cb) {
    if (curr) curr.kill();
    console.log('Turbo is running...');
    curr = exec('./turbo', function(err, stdout, stderr) {
    });
    cb();
});

gulp.task('test', function(cb) {
    exec('go test', function(err, stdout, stderr) {
        console.log(stdout);
        console.log(stderr);
        cb(err);
    });
});

gulp.task('watch', function () {
    gulp.watch('*.go', ['run']);
});

gulp.task('default', ['watch']);