module.exports = {
  // publicPath: process.env.NODE_ENV === 'production' ? '/scheduler/' : '/', // Used if service lives under a subpath
  devServer: {
    proxy: 'http://localhost:1337',
  }
}
