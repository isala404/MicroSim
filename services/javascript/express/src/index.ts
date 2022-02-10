import express from "express";
import yargs from 'yargs/yargs';
import { Route, Response } from "./model";
import { callNextDestination, castAndExcute } from "./utils"
import morgan from "morgan";

const argv = yargs(process.argv)
  .option('service-name', {
    required: true,
    default: 'Undefined',
    type: 'string',
    description: 'The name set on the response',
  })
  .option('port', {
    description: 'The port the web server will bind to',
    type: 'number',
    default: 9090,
  })
  .help()
  .alias('help', 'h')
  .parseSync();

const app = express()
app.use(express.json());
app.use((req, res, next) => {
  // tslint:disable-next-line:no-console
  console.log(`${new Date().toISOString()} RemoteAddr=${req.hostname} Method=${req.method} Path=${req.path} Body=${JSON.stringify(req.body)}`)
  next()
})

const port = argv.port;

app.post('/', async (req, res) => {
  const reply: Response = {
    service: argv.serviceName as string,
    address: "",
    errors: [],
    response: []
  };

  const payload: Route = { ...req.body }

  reply.address = payload.designation;

  for (const fault of payload.faults.before) {
    try {
      await castAndExcute(fault);
    } catch (error) {
      reply.errors.push(error.message);
    }
  }

  if (payload.routes) {
    for (const route of payload.routes) {
      let destRes: Response = null;
      try {
        destRes = await callNextDestination(route);
      } catch (e) {
        reply.errors.push(e.message);
      }
      reply.response.push(destRes);
    }
  }


  for (const fault of payload.faults.after) {
    try {
      await castAndExcute(fault);
    } catch (error) {
      reply.errors.push(error.message);
    }
  }

  res.send(reply);
})

app.listen(port, () => {
  // tslint:disable-next-line:no-console
  console.log(`App listening at http://localhost:${port}`)
})