.PHONY: lint test

lint:
	python3 -m pylint check_hwgroup

test:
	python3 -m unittest -v test_check_hwgroup.py
coverage:
	python3 -m coverage run -m unittest -b test_check_hwgroup.py
	python3 -m coverage report -m --include check_hwgroup.py
