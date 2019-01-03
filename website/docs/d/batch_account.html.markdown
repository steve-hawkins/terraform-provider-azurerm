---
layout: "azurerm"
page_title: "Azure Resource Manager: azurerm_batch_account"
sidebar_current: "docs-azurerm-datasource-batch-account"
description: |-
  Get information about an existing Batch Account

---

# Data Source: azurerm_batch_account

Use this data source to access information about an existing Batch Account.

## Example Usage

```hcl
data "azurerm_batch_account" "test" {
  name                = "testbatchaccount"
  resource_group_name = "test"
}

output "pool_allocation_mode" {
  value = "${data.azurerm_batch_account.test.pool_allocation_mode}"
}
```

## Argument Reference

* `name` - (Required) The name of the Batch account.

* `resource_group_name` - (Required) The Name of the Resource Group where this Batch account exists.

## Attributes Reference

The following attributes are exported:

* `id` - The Batch account ID.

* `name` - The Batch account name.

* `location` - The Azure Region in which this Batch account exists.

* `pool_allocation_mode` - The pool allocation mode configured for this Batch account.

* `storage_account_id` - The ID of the Storage Account used for this Batch account.

* `tags` - A map of tags assigned to the Batch account.
