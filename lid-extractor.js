const { Client } = require('pg');
const fs = require('fs');

async function extractLidMaster() {
    console.log("\n" + "╔" + "═".repeat(58) + "╗");
    console.log("║" + " ".repeat(18) + "💎 LID MASTER EXTRACTOR 💎" + " ".repeat(14) + "║");
    console.log("╚" + "═".repeat(58) + "╝");

    const client = new Client({
        connectionString: process.env.DATABASE_URL,
        ssl: { rejectUnauthorized: false }
    });

    try {
        await client.connect();
        console.log("✅ [DATABASE] Connected");

        const query = 'SELECT jid, lid FROM whatsmeow_device;';
        const res = await client.query(query);

        if (res.rows.length === 0) {
            console.log("⚠️ [EMPTY] No Session Found");
            process.exit(0);
        }

        console.log(`📊 [FOUND] all ${res.rows.length} session data received\n`);
        
        let botData = {};

        res.rows.forEach((row, index) => {
            if (row.jid && row.lid) {
                const purePhone = row.jid.split('@')[0].split(':')[0];
                const pureLid = row.lid.split('@')[0].split(':')[0] + "@lid";

                console.log(`  ╭────────────── [ BOT #${index + 1} ] ──────────────`);
                console.log(`  │ 📱 Phone ر : ${purePhone}`);
                console.log(`  │ 🆔  LID  : ${pureLid}`);
                console.log(`  │ ✨ status   : successfuly save`);
                console.log(`  ╰───────────────────────────────────────────\n`);

                botData[purePhone] = {
                    phone: purePhone,
                    lid: pureLid,
                    extractedAt: new Date().toISOString()
                };
            }
        });

        const finalJson = {
            timestamp: new Date().toISOString(),
            count: Object.keys(botData).length,
            bots: botData
        };

        fs.writeFileSync('./lid_data.json', JSON.stringify(finalJson, null, 2));
        console.log("💾 [SUCCESS] data 'lid_data.json' saved");

    } catch (err) {
        console.error("❌ [CRITICAL ERROR]:", err.message);
    } finally {
        await client.end();
        console.log("\n🏁 [FINISHED]۔");
        process.exit(0);
    }
}

extractLidMaster();