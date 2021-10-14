"use strict";

import {Fault} from '../model';
import crypto from 'crypto';

// creating memory leak of {size}MB for {duration} seconds
const createLeak = async (size: number, duration: number) => {
  // tslint:disable-next-line:no-console
  console.log(`creating memory leak of ${size}MB for ${duration} seconds`)

  let leak = crypto.randomBytes(size * 1024 * 1024);

  // Sleep for duration
  await new Promise(r => setTimeout(r, duration));

  leak = null;
  global.gc();

  // tslint:disable-next-line:no-console
  console.log(`memory leak cleared`)
}

class MemoryLeak implements Fault {
    type: string;
    args: any;

    constructor(args: any) {
        this.type = 'memory-leak';
        this.args = args;
    }

    async run(): Promise<void> {
        // tslint:disable-next-line:no-unused-expression
        new Promise(r => createLeak(this.args.size, this.args.duration));
        return;
    }
}

export default MemoryLeak;