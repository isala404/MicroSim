"use strict";
var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const express_1 = __importDefault(require("express"));
const yargs_1 = __importDefault(require("yargs/yargs"));
const utils_1 = require("./utils");
const argv = (0, yargs_1.default)(process.argv)
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
const app = (0, express_1.default)();
app.use(express_1.default.json());
const port = argv.port;
app.post('/', (req, res) => __awaiter(void 0, void 0, void 0, function* () {
    const reply = {
        service: argv.serviceName,
        address: "",
        errors: [],
        response: []
    };
    const payload = Object.assign({}, req.body);
    reply.address = payload.designation;
    for (const fault of payload.faults.before) {
        try {
            yield (0, utils_1.castAndExcute)(fault);
        }
        catch (error) {
            reply.errors.push(error.message);
        }
    }
    if (payload.routes) {
        for (const route of payload.routes) {
            let destRes = null;
            try {
                destRes = yield (0, utils_1.callNextDestination)(route);
            }
            catch (e) {
                reply.errors.push(e.message);
            }
            reply.response.push(destRes);
        }
    }
    for (const fault of payload.faults.after) {
        try {
            yield (0, utils_1.castAndExcute)(fault);
        }
        catch (error) {
            reply.errors.push(error.message);
        }
    }
    res.send(reply);
}));
app.listen(port, () => {
    // tslint:disable-next-line:no-console
    console.log(`App listening at http://localhost:${port}`);
});
//# sourceMappingURL=index.js.map