
export type TestRequest = {
    test: string
}

export type TestResponse = {
    test: string
}

export type ModuleInterface = {
    actions: {
        test: {
            request_type: TestRequest,
            response_type: TestResponse,
        }
    }
}