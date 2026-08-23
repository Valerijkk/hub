import {
	QueryClient,
	QueryObserver,
	type QueryObserverResult,
} from "@tanstack/react-query";

let querySequence = 0;

function nextQueryKey(): readonly [string, number] {
	querySequence += 1;
	return ["test-query-result", querySequence];
}

export function querySuccess<TData, TError = Error>(
	data: TData,
): QueryObserverResult<TData, TError> {
	const client = new QueryClient();
	const queryKey = nextQueryKey();
	client.setQueryData(queryKey, data);
	return new QueryObserver<TData, TError>(client, {
		queryKey,
	}).getCurrentResult();
}

export function queryError<TData = unknown, TError = Error>(
	error: TError,
	refetch?: QueryObserverResult<TData, TError>["refetch"],
): QueryObserverResult<TData, TError> {
	const client = new QueryClient();
	const queryKey = nextQueryKey();
	const query = client.getQueryCache().build<TData, TError>(client, {
		queryKey,
	});
	query.setState({
		error,
		status: "error",
		fetchStatus: "idle",
	});
	const result = new QueryObserver<TData, TError>(client, {
		queryKey,
	}).getCurrentResult();
	return refetch ? { ...result, refetch } : result;
}
