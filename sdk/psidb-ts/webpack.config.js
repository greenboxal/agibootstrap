const path = require('path');
const fs = require("fs");

const appDirectory = fs.realpathSync(process.cwd());
const resolveApp = relativePath => path.resolve(appDirectory, relativePath);

const paths = {
    appPath: resolveApp('.'),
    appPackageJson: resolveApp('package.json'),
    appSrc: resolveApp('src'),
    appTsConfig: resolveApp('tsconfig.json'),
    appJsConfig: resolveApp('jsconfig.json'),
    appNodeModules: resolveApp('node_modules'),
};

const moduleFileExtensions = [
    'web.mjs',
    'mjs',
    'web.js',
    'js',
    'web.ts',
    'ts',
    'web.tsx',
    'tsx',
    'json',
    'web.jsx',
    'jsx',
];

module.exports = {
    mode: 'development',
    target: 'es2020',
    devtool: 'source-map',

    optimization: {
        runtimeChunk: 'single',

        splitChunks: {
            chunks: 'all',
            maxInitialRequests: Infinity,
            minSize: 0,
            cacheGroups: {
                vendor: {
                    test: /[\\/]node_modules[\\/]/,
                    name(module) {
                        // get the name. E.g. node_modules/packageName/not/this/part.js
                        // or node_modules/packageName
                        const packageName = module.context.match(/[\\/]node_modules[\\/](.*?)([\\/]|$)/)[1];

                        // npm package names are URL-safe, but some servers don't like @ symbols
                        return `npm.${packageName.replace('@', '')}`;
                    },
                },
            },
        },
    },

    entry: {
        main: {
            import: './src/index.ts',
            library: {
                type: 'commonjs-module'
            }
        },
    },

    module: {
        strictExportPresence: true,

        rules: [
            {
                enforce: 'pre',
                exclude: /@babel(?:\/|\\{1,2})runtime/,
                test: /\.(js|mjs|jsx|ts|tsx|css)$/,
                loader: require.resolve('source-map-loader'),
            },

            {
                test: /\.(js|mjs|jsx|ts|tsx)$/,
                include: paths.appSrc,
                loader: require.resolve('babel-loader'),
                options: {
                    customize: require.resolve(
                        'babel-preset-react-app/webpack-overrides'
                    ),
                    presets: [
                        [
                            require.resolve('babel-preset-react-app'),
                            {
                                runtime: 'automatic'
                            },
                        ],
                    ],

                    // This is a feature of `babel-loader` for webpack (not Babel itself).
                    // It enables caching results in ./node_modules/.cache/babel-loader/
                    // directory for faster rebuilds.
                    cacheDirectory: true,
                    // See #6846 for context on why cacheCompression is disabled
                    cacheCompression: false,
                    compact: false,
                },
            },
            // Process any JS outside of the app with Babel.
            // Unlike the application JS, we only compile the standard ES features.
            {
                test: /\.(js|mjs)$/,
                exclude: /@babel(?:\/|\\{1,2})runtime/,
                loader: require.resolve('babel-loader'),
                options: {
                    babelrc: false,
                    configFile: false,
                    compact: false,
                    presets: [
                        [
                            require.resolve('babel-preset-react-app/dependencies'),
                            { helpers: true },
                        ],
                    ],
                    cacheDirectory: true,
                    // See #6846 for context on why cacheCompression is disabled
                    cacheCompression: false,

                    // Babel sourcemaps are needed for debugging into node_modules
                    // code.  Without the options below, debuggers like VSCode
                    // show incorrect code and set breakpoints on the wrong lines.
                    sourceMaps: true,
                    inputSourceMap: true,
                },
            },
        ],
    },

    resolve: {
        modules: ['node_modules'],
        extensions: moduleFileExtensions,
    },

    output: {
        filename: '[name]',
        chunkFilename: '[id].js',
        chunkFormat: 'commonjs',
        chunkLoading: 'require',
        path: path.resolve(__dirname, 'dist'),
    },

    experiments: {
        outputModule: true,
    }
};
