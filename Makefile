.PHONY: lint test

lint:
	python -m pylint check_hwgroup

test:
	python -m unittest -v test_check_hwgroup.py
coverage:
	python -m coverage run -m unittest test_check_hwgroup.py
	python -m coverage report -m --include check_hwgroup.py
