run "setup_tests" {
    module {
        source = "./tests/setup"
    }
}

run "create_bucket" {
  command = apply

  variables {
    dyndns_comment = "terraformcomment"
    dyndns_password = "terraformpassword"
    dyndns_zone = "mahn.ke"
    dyndns_label = "terraformzone.by.vincent"
    dyndns_target_ip = "1.2.3.4"
  }

  # Check that the bucket name is correct
    assert {
        condition     = can(regex("^dyn.*[0-9]+$", run.setup_tests.ddns_login))
        error_message = "ddns_login must start with 'dyn' and end with at least one number"
    }
}
