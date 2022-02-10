import { Route, Response, Fault } from "./model"
import fetch from 'cross-fetch';
import Latency from './faults/latency';
import MemoryLeak from './faults/memory-leak';


export const callNextDestination = async (route: Route): Promise<Response> => {
  // tslint:disable-next-line:no-console
  console.log(`Calling Next Destination, Designation=${route.designation} Body=${JSON.stringify(route)}`)
  const rawResponse = await fetch(route.designation, {
    method: 'POST',
    headers: {
      'Accept': 'application/json',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(route)
  });
  const response = await rawResponse.json() as Response;

  return response;
}

// Not sure if this the right way to do this but hey, it works :)
export const castAndExcute = async (fault: Fault) => {
  let rawFault: any;

  switch (fault.type) {
    case 'latency':
      rawFault = new Latency(fault.args);
      break;
    case 'memory-leak':
      rawFault = new MemoryLeak(fault.args);
      break;
    default:
      throw new Error(`Fault type ${fault.type} not implemented`);
  }

  await rawFault.run();
}