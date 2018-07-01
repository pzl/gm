const StyleLintPlugin = require('stylelint-webpack-plugin')

module.exports = {
  env: {
  	api: process.env.API || "http://localhost:8080"
  },

  /*
  ** Headers of the page
  */
  head: {
    title: 'manager',
    meta: [
      { charset: 'utf-8' },
      { 'http-equiv': 'x-ua-compatible', content: 'ie=edge' },
      { name: 'viewport', content: 'width=device-width, initial-scale=1' },
      { hid: 'description', name: 'description', content: 'Manage resources and services' }
    ],
    link: [
      { rel: 'icon', type: 'image/x-icon', href: '/favicon.ico' }
    ]
  },
  /*
  ** Customize the progress bar color
  */
  loading: { color: '#3B8070' },
  css: [
  	'@/assets/css/normalize.css',
  	'@/assets/css/main.css',
  ],
  /*
  ** Build configuration
  */
  build: {
  	plugins: [
  		new StyleLintPlugin({
  			configFile: '.stylelintrc',
  			files: ['assets/css/main.css', '**/*.vue']
  		})
  	],
  	extractCSS: true,
  	postcss: {
  		plugins: {
  			'postcss-import': {
  				//path: ['assets/css']
  			},
  			'postcss-cssnext': {
  				'browsers': ['last 2 versions']
  			},
  			'postcss-url': {},
  			'postcss-pxtorem': {}
  		}
  	},
  	vendor: [
  		'axios',
  	],

    /*
    ** Run ESLint on save
    */
    extend (config, { isDev, isClient }) {
      config.devtool = false

      if (isDev && isClient) {
        config.module.rules.push({
          enforce: 'pre',
          test: /\.(js|vue)$/,
          loader: 'eslint-loader',
          exclude: /(node_modules)/
        })
      }

      let vueLoader = config.module.rules.find(rule => rule.loader === 'vue-loader');
      vueLoader.options.transformToRequire = {
        img: 'src',
        image: 'xlink:href',
        use: 'href',
        use: 'xlink:href',
        object: 'data',
        video: 'src',
        source: 'src'
      }

      let fileLoader = config.module.rules.find(rule => rule.test.toString().includes('svg'))
      fileLoader.test = /\.(png|jpe?g|gif)$/

      config.module.rules.push({
        test: /\.svg$/,
        exclude: /node_modules/,
        oneOf: [
          {
            resourceQuery: /inline/,
            loader: 'vue-svg-loader',
            options: {
              svgo: {
                plugins: [
                  {removeViewBox: false},
                  {removeUselessStrokeAndFill: false}
                ]
              }
            }
          },
          {
            use: [
              {
                loader: 'svg-url-loader',
                options: {
                  name: 'img/[name].[hash:7].[ext]',
                  limit: 1000,
                  stripdeclarations: true,
                  iesafe: true,
                  noquotes: true
                }
              },
              'svg-fill-loader'
            ]

          }
        ]
      })

    }
  },
  watchers: {
    chokidar: {
      usePolling: false,
      ignored: '*node_modules*'
    }
  }
}
