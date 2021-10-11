"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const express_1 = __importDefault(require("express"));
const yargs_1 = __importDefault(require("yargs/yargs"));
const argv = (0, yargs_1.default)(process.argv)
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
const app = (0, express_1.default)();
const port = argv.port;
app.get('/', (req, res) => {
    res.send('Hello World!');
});
app.listen(port, () => {
    // tslint:disable-next-line:no-console
    console.log(`Example app listening at http://localhost:${port}`);
});
//# sourceMappingURL=index.js.map