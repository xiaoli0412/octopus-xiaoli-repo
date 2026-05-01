const React = require('../web/node_modules/react');

const SelectContext = React.createContext({
	items: [],
	onValueChange: undefined,
	value: undefined,
});

function collectItems(children, items = []) {
	React.Children.forEach(children, (child) => {
		if (!React.isValidElement(child)) return;
		if (child.type && child.type.__verifyIsSelectItem) {
			items.push({
				label: typeof child.props.children === 'string' ? child.props.children : String(child.props.value),
				value: child.props.value,
			});
			return;
		}
		if (child.props && child.props.children) {
			collectItems(child.props.children, items);
		}
	});
	return items;
}

function Select({ children, onValueChange, value }) {
	const items = React.useMemo(() => collectItems(children, []), [children]);
	const contextValue = React.useMemo(() => ({ items, onValueChange, value }), [items, onValueChange, value]);
	return React.createElement(SelectContext.Provider, { value: contextValue }, children);
}

function SelectTrigger({ className }) {
	const context = React.useContext(SelectContext);
	return React.createElement(
		'select',
		{
			className,
			onChange: (event) => context.onValueChange?.(event.target.value),
			role: 'combobox',
			value: context.value,
		},
		context.items.map((item) => React.createElement('option', { key: item.value, value: item.value }, item.label)),
	);
}

function SelectValue() {
	return null;
}

function SelectContent() {
	return null;
}

function SelectItem() {
	return null;
}

SelectItem.__verifyIsSelectItem = true;

module.exports = {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
};
