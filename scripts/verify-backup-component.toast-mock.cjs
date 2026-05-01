function getState() {
	return global.__backupComponentVerifyState;
}

module.exports = {
	toast: {
		success(message) {
			getState().toastSuccessCalls.push(message);
		},
		error(message) {
			getState().toastErrorCalls.push(message);
		},
	},
};
