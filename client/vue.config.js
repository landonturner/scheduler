module.exports = {
  publicPath: process.env.NODE_ENV === 'production' ? '/scheduler/' : '/',
  devServer: {
    proxy: 'http://localhost:1337',
  }
}
