const path = require('path')
const MonacoWebpackPlugin = require('monaco-editor-webpack-plugin')

module.exports = {
    runtimeCompiler: true,
    lintOnSave: false,
    css: {
        extract: process.env.NODE_ENV === 'production',
        sourceMap: process.env.NODE_ENV === 'production',
        loaderOptions: {
            less: {
                relativeUrls: true,
                paths: [
                    path.join(__dirname, './src'),
                    path.join(__dirname, './node_modules')
                ]
            }
        }
    },

    chainWebpack: config => {
        config.module
            .rule('md')
            .test(/\.md$/)
            .end()
        config.module
            .rule('proto')
            .test(/\.proto/)
            .use('proto-loader')
            .loader('proto-loader')
            .end()
        config.module
            .rule('vue')
            .test(/\.vue/)
            .use('vue-loader')
            .loader('vue-loader')
            .end()
    },
    configureWebpack: {
        // Set up all the aliases we use in our app.
        resolve: {
            alias: {
                'vuejs-datatable': 'vuejs-datatable/dist/vuejs-datatable.esm.js'
            }
        },
        plugins: [new MonacoWebpackPlugin({
            filename: "js/[name].worker.js"
        })]
    },

    devServer: {
        port: 8081,
        proxy: {
            '/api': {
                target: 'http://localhost:8080/api',
                changeOrigin: true,
                pathRewrite: {
                    '^/api': ''
                }
            }
        }
    }
}
