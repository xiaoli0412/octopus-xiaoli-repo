module.exports = {
	useTranslations() {
		return (key) => {
			if (key === 'common.helpHint.ariaLabel' || key === 'ariaLabel') {
				return 'View help';
			}
			return key;
		};
	},
};
