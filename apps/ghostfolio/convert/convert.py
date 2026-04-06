import json
import csv
from datetime import datetime


def is_regular_transaction(transaction):
    return transaction["type"] == "SECURITY_TRANSACTION" and transaction["status"] == "SETTLED"


def is_transfer_in(transaction):
    return transaction["type"] == "NON_TRADE_SECURITY_TRANSACTION" and transaction["nonTradeSecurityTransactionType"] == "TRANSFER_IN"


def is_dividend(transaction):
    return transaction["type"] == "CASH_TRANSACTION" and transaction["cashTransactionType"] == "DISTRIBUTION"


def is_deposit(transaction):
    return transaction["type"] == "CASH_TRANSACTION" and transaction["cashTransactionType"] == "DEPOSIT"


def is_tax(transaction):
    return transaction["type"] == "CASH_TRANSACTION" and transaction["cashTransactionType"] == "TAX"


def is_swap_in(transaction):
    return transaction["type"] == "NON_TRADE_SECURITY_TRANSACTION" and transaction["nonTradeSecurityTransactionType"] == "SWAP_IN"


def is_swap_out(transaction):
    return transaction["type"] == "NON_TRADE_SECURITY_TRANSACTION" and transaction["nonTradeSecurityTransactionType"] == "SWAP_OUT"


def create_csv_from_transactions(input_file, output_file):
    # Read JSON data from file
    with open(input_file, 'r') as file:
        transactions = json.load(file)

    with open(output_file, mode='w', newline='') as file:
        writer = csv.writer(file)

        # Write CSV header
        writer.writerow(["Date", "Code", "DataSource", "Currency", "Price", "Quantity", "Action", "Fee", "Note", "AccountId"])

        for transaction in transactions["transactions"]:
            if is_deposit(transaction):
                continue

            try:
                date = datetime.strptime(transaction["lastEventDateTime"], "%Y-%m-%dT%H:%M:%S.%fZ").strftime("%d-%m-%Y")
            except ValueError:
                date = datetime.strptime(transaction["lastEventDateTime"], "%Y-%m-%dT%H:%M:%SZ").strftime("%d-%m-%Y")

            currency = transaction["currency"]
            data_source = "MANUAL"
            code = transaction["isin"] if "isin" in transaction else transaction["relatedIsin"]

            # if not isin:
            #     pass
            # elif isin == "IE00B5SSQT16":
            #     code = "H410.DE"
            # elif isin == "LU1737652583":
            #     code = "AEMD.DE"
            # elif isin == "LU1737652237":
            #     code = isin
            #     data_source = "MANUAL"
            if code == "US0378331005":
                data_source = "YAHOO"
                code = "AAPL"
                currency = "USD"  # TODO 168,5056
            elif code == "AT0000857164":
                continue
                data_source = "YAHOO"
                code = code + ".SG"
            # else:
            #     code = isin + ".SG"

            note = ""
            accountId = "1b4f056d-65cb-46e6-ac0a-7a6393a5c2c2"

            if is_regular_transaction(transaction):
                quantity = transaction["quantity"]
                price = abs(transaction["amount"] / quantity)
                fee = 0.99 if transaction["securityTransactionType"] == "SINGLE" else 0
                action = transaction["side"].lower()
                writer.writerow([date, code, data_source, currency, price, quantity, action, fee, note, accountId])

            elif is_transfer_in(transaction) or is_swap_in(transaction):
                quantity = transaction["quantity"]
                price = abs(transaction["amount"] / quantity)
                fee = 0
                action = "buy"
                note = "transfer in"
                writer.writerow([date, code, data_source, currency, price, quantity, action, fee, note, accountId])

            elif is_swap_out(transaction):
                quantity = transaction["quantity"]
                price = abs(transaction["amount"] / quantity)
                fee = 0
                action = "sell"
                note = "transfer out"
                writer.writerow([date, code, data_source, currency, price, quantity, action, fee, note, accountId])

            elif is_dividend(transaction):
                quantity = 1
                price = transaction["amount"]  # TODO: divide by count
                fee = 0
                action = "dividend"
                writer.writerow([date, code, data_source, currency, price, quantity, action, fee, note, accountId])

            # elif is_tax(transaction):
            #     quantity = 0
            #     price = 0
            #     fee = transaction["amount"]
            #     action = "fee"
            #     note = "tax"
            #     writer.writerow([date, code, data_source, currency, price, quantity, action, fee, note, accountId])


# Specify input and output file paths
input_file = 'scalable.json'
output_file = 'ghostfolio.csv'

# Call the function to create CSV
create_csv_from_transactions(input_file, output_file)