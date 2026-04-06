import csv
from datetime import datetime


def create_csv_from_transactions(input_file, output_file):
    # Read JSON data from file
    fnzfile = open(input_file, 'r')

    with open(output_file, mode='w', newline='') as file:
        writer = csv.writer(file)

        # Write CSV header
        writer.writerow(["Date", "Code", "DataSource", "Currency", "Price", "Quantity", "Action", "Fee", "Note", "AccountId"])

        reader = csv.DictReader(fnzfile, delimiter=';')
        for row in reader:
            date = row['Buchungsdatum']
            isin = row['ISIN']
            betrag = int(row['Zahlungsbetrag'])
            quantity = float(row['Anteile'].replace(',', '.'))

            date = datetime.strptime(date, "%d.%m.%y").strftime("%d-%m-%Y")

            currency = "EUR"
            data_source = "MANUAL"
            code = isin

            note = ""
            accountId = "a2ff6e71-4427-427c-95e3-a66e11841f2d"

            if betrag > 0:
                price = abs(betrag / quantity)
                fee = 0
                action = "buy"
                writer.writerow([date, code, data_source, currency, price, quantity, action, fee, note, accountId])

            else:
                quantity = quantity
                price = abs(3 / quantity)
                fee = 0
                action = "sell"
                note = "3€ fee"
                writer.writerow([date, code, data_source, currency, price, quantity, action, fee, note, accountId])


# Specify input and output file paths
input_file = 'fnz.csv'
output_file = 'ghostfolio_vwl.csv'

# Call the function to create CSV
create_csv_from_transactions(input_file, output_file)