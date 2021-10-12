import { Route, Response } from "./model"
import fetch from 'cross-fetch';


export const callNextDestination = async (route: Route): Promise<Response> => {

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