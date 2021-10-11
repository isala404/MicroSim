import express from "express";
import yargs from 'yargs/yargs';

const argv = yargs(process.argv)
.option('service-name', {
  required: true,
  default: 'Undefined',
  description: 'The name set on the response',
})
.option('port', {
    description: 'The port the web server will bind to',
    type: 'number',
    default: 3000,
})
.help()
.alias('help', 'h')
.parseSync();

const app = express()
const port = argv.port;

app.get('/', (req, res) => {
  res.send('Hello World!')
})

app.listen(port, () => {
  // tslint:disable-next-line:no-console
  console.log(`Example app listening at http://localhost:${port}`)
})