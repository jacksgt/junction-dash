import serial
import time
import psycopg2
'''
CREATE TABLE wifi_diy (
   ts timestamp without time zone default (now() at time zone 'utc'),
   station TEXT,
   rssi INTEGER,
   mac TEXT
);
INSERT INTO wifi_diy(timestamp, station, rssi, mac)
VALUES
   (
    '1573899334.94',
    'TEMPID',
    '-81',
    'a0:c5:89:a5:55:a5'
   );

SELECT * 
FROM wifi_diy
WHERE extract(epoch from wifi_diy.ts) >= 1573900210
AND extract(epoch from wifi_diy.ts) < 1573902110;


select extract(epoch from wifi_diy.ts::timestamp) as ts FROM wifi_diy;
'''

conn = psycopg2.connect(dbname='defaultdb', user='avnadmin', password='yqt245bcfy6o5xmo', host='pg-6ef61e2-aalto-2dd2.aivencloud.com', port='28694', sslmode='require')
cursor = conn.cursor()
   


with serial.Serial('/dev/cu.SLAB_USBtoUART', 115200) as ser:
    while(True):
        line = ser.readline()
        line = line.strip("\r\n")
        print(line)
#        timestamp = str(time.time())
        station,rssi,mac = line.split(",")
        cursor.execute("INSERT INTO wifi_diy (station, rssi, mac) VALUES(%s, %s, %s)", (station,rssi,mac))
        conn.commit()
