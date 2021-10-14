import {Fault} from '../model';

class Latency implements Fault {
    type: string;
    args: any;

    constructor(args: any) {
        this.type = 'latency';
        this.args = args;
    }

    async run(): Promise<void> {
        await new Promise(r => setTimeout(r, this.args.delay));
        return;
    }
}

export default Latency;