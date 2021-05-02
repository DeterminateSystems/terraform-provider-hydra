##### Description

<!--- Please include a short description of what your PR does and / or the
motivation behind it --->

##### Checklist

<!--- Use `nix-shell` for a shell with all the required dependencies for
building / formatting / testing / etc. --->

- [ ] Built with `make build`
- [ ] Formatted with `make fmt`
- [ ] Verifed the example configuration still parses with `terraform init && terraform validate` (you may need to `make install` from the root of the project)
- [ ] Ran acceptance tests with `HYDRA_HOST=http://0:63333 HYDRA_USERNAME=alice HYDRA_PASSWORD=foobar make testacc` (you will need to spin up a local / temporary instance of Hydra)
- [ ] Added or updated relevant documentation (leave unchecked if not applicable)
