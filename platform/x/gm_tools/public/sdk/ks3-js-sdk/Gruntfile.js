
module.exports = function(grunt) {
  // Project configuration.
  grunt.util.linefeed = '\n';

  grunt.initConfig({
    pkg: grunt.file.readJSON('package.json'),
    dist: 'dist',
    src: 'src',
    filename: 'ks3jssdk',
    uglify: {
      options: {
        banner: '/*! <%= filename %> <%= grunt.template.today("dd-mm-yyyy") %> */\n'
      },
      dist:{
        src:['src/<%= filename %>.js'],
        dest:'<%= dist %>/<%= filename %>.min.js'
      }
    },
    copy: {
     copy2demo: {
        files: [
          {expand: true,
          src: ['<%= filename %>.min.js'],
          cwd:  '<%= dist %>/',
          dest: 'demo/js/'},

          {
            expand: true,
            src: ['<%= filename %>.js'],
            cwd:  '<%= src %>/',
            dest: 'demo/js/'
          }
        ]
      }
    },
    jshint: {
      files: ['Gruntfile.js','src/ks3jssdk.js']
    }
  });

  grunt.loadNpmTasks('grunt-contrib-uglify');
  grunt.loadNpmTasks('grunt-contrib-jshint');
  grunt.loadNpmTasks('grunt-contrib-copy');

  grunt.registerTask('default', ['uglify', 'copy']);

  return grunt;
};
