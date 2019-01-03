# Configure the Microsoft Azure Provider
provider "azurerm" {
  # if you're using a Service Principal (shared account) then either set the environment variables, or fill these in: 
  # subscription_id = "..."
  # client_id       = "..." 
  # client_secret   = "..."
  # tenant_id       = "..."
}

resource "azurerm_resource_group" "rg" {
  name     = "${var.resource_group_name}"
  location = "${var.location}"
}

resource "random_integer" "ri" {
  min = 10000
  max = 99999
}

resource "azurerm_storage_account" "stor" {
  name                     = "stor${random_integer.ri.result}"
  resource_group_name      = "${azurerm_resource_group.rg.name}"
  location                 = "${azurerm_resource_group.rg.location}"
  account_tier             = "${var.storage_account_tier}"
  account_replication_type = "${var.storage_replication_type}"
}

resource "azurerm_batch_account" "batch" {
  name                     = "batch${random_integer.ri.result}"
  resource_group_name      = "${azurerm_resource_group.rg.name}"
  location                 = "${azurerm_resource_group.rg.location}"
  storage_account_name     = "${azurerm_storage_account.stor.name}"
}