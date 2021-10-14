export interface Fault {
    type: string;
    args: any;
    run(): Promise<void>;
}

export interface FaultState {
    before: Fault[];
    after: Fault[];
}

export interface Route {
    designation: string;
    faults: FaultState;
    routes?: Route[];
}

export interface Response {
    service: string;
    address: string;
    errors: string[];
    response: Response[];
}
