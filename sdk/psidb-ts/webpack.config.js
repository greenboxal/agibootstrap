const fs = require("fs");
const webpack = require("webpack");
const path = require('path');
const { WebpackManifestPlugin } = require('webpack-manifest-plugin');
const ModuleNotFoundPlugin = require('react-dev-utils/ModuleNotFoundPlugin');
const ESLintPlugin = require('eslint-webpack-plugin');
const DeclarationBundlerPlugin = require('types-webpack-bundler');
const tstReflectTransform = require("tst-reflect-transformer").default;

const shouldUseSourceMap = process.env.GENERATE_SOURCEMAP !== 'false';
const emitErrorsAsWarnings = process.env.ESLINT_NO_DEV_ERRORS === 'true';

const appDirectory = fs.realpathSync(process.cwd());
const resolveApp = relativePath => path.resolve(appDirectory, relativePath);

const hasJsxRuntime = (() => {
    if (process.env.DISABLE_NEW_JSX_TRANSFORM === 'true') {
        return false;
    }

    try {
        require.resolve('react/jsx-runtime');
        return true;
    } catch (e) {
        return false;
    }
})();

const paths = {
    appPath: resolveApp('.'),
    appPackageJson: resolveApp('package.json'),
    appSrc: resolveApp('src'),
    appTsConfig: resolveApp('tsconfig.json'),
    appJsConfig: resolveApp('jsconfig.json'),
    appNodeModules: resolveApp('node_modules'),
    appTsBuildInfoFile: resolveApp('node_modules/.cache/tsconfig.tsbuildinfo'),
    publicUrlOrPath: "QmYXZ//",
};

const moduleFileExtensions = [
    '.web.mjs',
    '.mjs',
    '.web.js',
    '.js',
    '.web.ts',
    '.ts',
    '.web.tsx',
    '.tsx',
    '.json',
    '.web.jsx',
    '.jsx',
];

const PSIDB_APP_ = /^PSIDB_APP_/i;

function getClientEnvironment() {
    const raw = Object.keys(process.env)
        .filter(key => PSIDB_APP_.test(key))
        .reduce(
            (env, key) => {
                env[key] = process.env[key];
                return env;
            },
            {
                // Useful for determining whether we’re running in production mode.
                // Most importantly, it switches React into the correct mode.
                NODE_ENV: process.env.NODE_ENV || 'development',
            }
        );
    // Stringify all values so we can feed into webpack DefinePlugin
    const stringified = {
        'process.env': Object.keys(raw).reduce((env, key) => {
            env[key] = JSON.stringify(raw[key]);
            return env;
        }, {}),
    };

    return { raw, stringified };
}

module.exports = function(webpackEnv) {
    const isEnvDevelopment = webpackEnv === 'development';
    const isEnvProduction = webpackEnv === 'production';
    const env = getClientEnvironment()

    return {
        mode: 'development',
        target: 'es2020',
        devtool: 'source-map',

        plugins: [
            // This gives some necessary context to module not found errors, such as
            // the requesting resource.
            new ModuleNotFoundPlugin(paths.appPath),
            new webpack.DefinePlugin(env.stringified),
            new WebpackManifestPlugin({
                fileName: 'asset-manifest.json',
                publicPath: paths.publicUrlOrPath,
                generate: (seed, files, entrypoints) => {
                    const manifestFiles = files.reduce((manifest, file) => {
                        manifest[file.name] = file.path;
                        return manifest;
                    }, seed);

                    const entrypointFiles = entrypoints.main.filter(
                        fileName => !fileName.endsWith('.map')
                    );

                    return {
                        files: manifestFiles,
                        entrypoints: entrypointFiles,
                    };
                },
            }),
            new ESLintPlugin({
                // Plugin options
                extensions: ['js', 'mjs', 'jsx', 'ts', 'tsx'],
                formatter: require.resolve('react-dev-utils/eslintFormatter'),
                eslintPath: require.resolve('eslint'),
                failOnError: !(isEnvDevelopment && emitErrorsAsWarnings),
                context: paths.appSrc,
                cache: true,
                cacheLocation: path.resolve(
                    paths.appNodeModules,
                    '.cache/.eslintcache'
                ),
                // ESLint class options
                cwd: paths.appPath,
                resolvePluginsRelativeTo: __dirname,
                baseConfig: {
                    extends: [require.resolve('eslint-config-react-app/base')],
                    rules: {
                        ...(!hasJsxRuntime && {
                            'react/react-in-jsx-scope': 'error',
                        }),
                    },
                },
            }),

            new DeclarationBundlerPlugin({
                moduleName: "x.y.z",
                out:'bundle.d.ts',
            }),
        ],

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
            ModuleExport: {
                import: './src/export.ts',
            }
        },

        module: {
            strictExportPresence: true,

            rules: [
                {
                    enforce: 'pre',
                    exclude: /@babel(?:\/|\\{1,2})runtime/,
                    test: /\.(js|mjs|jsx|ts|tsx)$/,
                    loader: require.resolve('source-map-loader'),
                },
                {
                    test: /\.ts(x?)$/,
                    include: paths.appSrc,
                    loader: require.resolve('ts-loader'),
                    options: {
                        getCustomTransformers: (program) => ({
                            before: [
                                tstReflectTransform(program, {})
                            ]
                        })
                    }
                },
                {
                    test: /\.(js|mjs|jsx)$/,
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
                                {helpers: true},
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
            alias: {

            }
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
}
