我是光年实验室高级招聘经理。
我在github上访问了你的开源项目，你的代码超赞。你最近有没有在看工作机会，我们在招软件开发工程师，拉钩和BOSS等招聘网站也发布了相关岗位，有公司和职位的详细信息。
我们公司在杭州，业务主要做流量增长，是很多大型互联网公司的流量顾问。公司弹性工作制，福利齐全，发展潜力大，良好的办公环境和学习氛围。
公司官网是http://www.gnlab.com,公司地址是杭州市西湖区古墩路紫金广场B座，若你感兴趣，欢迎与我联系，
电话是0571-88839161，手机号：18668131388，微信号：echo 'bGhsaGxoMTEyNAo='|base64 -D ,静待佳音。如有打扰，还请见谅，祝生活愉快工作顺利。

# Scheduler

This repo allows authenticated users to create a schedule which hits http
endpoints via a web interface. The web interface is built with vue and
vuetify and the backend api is a golang server. It uses a simple sqlite
database as its persistance layer and JWT for auth.

The go server handles the api endpoints and also serves the static website
files. When a specific file is not located `http://localhost:1337/asdf`,
rather than the server delivering an asset or a 404, it sends the main
index.html page and lets vue handle the 404 page. This allows us to use
"history mode" with vue.

## Usages

This can be used to schedule http calls to any remote endpoint. This can be
very useful for webhook endpoints. Scheduling jenkins builds or netlify deploys
or slack messages through the webook interface are just a few examples.

## Development Instructions

In production the api server serves the client files, but when in development
it is more convenient to use vue's development server so you can leverage the
watching functionality (auto reload browser when a code change has happened).
To do this, start both the api and the client web server. Connect to the front
end through your favorite browser to http://localhost:8080 (as instructed in
your console). CORS is not an issue despite the two services running on
different ports because the vue-cli-service proxys the api server (configured
in client/vue.config.js).

#### Starting the API Server

Start the api server in the standard way. 
```
make run
```

#### Starting the Client

Navigate to the client folder.

Install all dependencies with
```
npm install
```

Run the following to serve the files for dev
```
npm run serve
```

This will watch the client folder for changes and auto reload in your browser.

## Production Instructions

For production, we want to build the client files for the server then just
start the server. I have added a dockerfile if that suits your fancy.

#### Building the Client
Build the client so that the golang server has files to serve. From the
client folder run
```
npm run build
```

#### Running the Server
Start it however you would usally start a go proj (build the binary and run
that or simply go run the main file). Provide the necessary environment
variables like so 

```
env JWT_SECRET="im a 32 bit hex encoded secret" REMOTE_URL="https://example.com/endpoint" go run server.go
```

*Note the persistence comes in the form of a file named `sqlite.db` that is
located in the folder you are running the program. This is pretty sloppy,
should be configurable and is a great target for a next step if anyone even
bothers to look at this.

#### Docker
The Dockerfile uses multi stage builds to minimize final image size. It builds
all the client code in a node container, and the golang code in a golang
container then mounts the two into an alpine container. The resulting size was
22.7MB. The default image name is `scheduler`.

```
make build
make docker-run
```

## Registering Users

Right now, there is no front end page to register a new user, so this curl
command is the easiest way to do it (anything interacting with the api would
work honestly, there are no checks to validate the source of the register
route). If there is any traction on this repo I'll consider putting a real user
management experience in the UI. Note here, if you are using special characters
you're going to have to encode them.

```
curl localhost:1337/register -d 'email=me@email.com&password=123&name=ME'
```

## Helpful Curl Commands to the API

```
curl localhost:1337/register -d 'email=me@email.com&password=123&name=ME'
JWT=$(curl localhost:1337/login -d 'email=me@email.com&password=123')
curl -H "Authorization: Bearer $JWT" localhost:1337/me
curl -H "Authorization: Bearer $JWT" localhost:1337/schedules 
curl -H "Authorization: Bearer $JWT" localhost:1337/schedules -d 'time=2002-10-02T10:00:00-05:00'
```
