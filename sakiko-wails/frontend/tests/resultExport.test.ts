import { describe, expect, it } from "vitest";
import { buildExportSections } from "../src/utils/resultExport";

describe("buildExportSections privacy masking", () => {
  it("masks CN inbound ASN, IP, and info in topology rows", () => {
    const archive = {
      version: 1,
      task: {
        id: "task-1",
        vendor: "mihomo",
        context: {
          profileName: "Example",
          preset: "full",
        },
      },
      state: {},
      results: [
        {
          proxyInfo: {
            name: "HK Node 01",
            type: "shadowsocks",
            address: "z.h.s.1.p.i.b.e.d.hk05-ae5.entry.v51124-5.qpon",
          },
          matrices: [
            {
              type: "GEOIP_INBOUND",
              payload: {
                address: "z.h.s.1.p.i.b.e.d.hk05-ae5.entry.v51124-5.qpon",
                ip: "114.28.148.163",
                asn: 4811,
                asOrganization: "Shanghai Information Network CO.,LTD.",
                isp: "China Telecom group",
                country: "China",
                city: "Hongkou",
                countryCode: "CN",
              },
            },
            {
              type: "GEOIP_OUTBOUND",
              payload: {
                ip: "103.151.173.98",
                asn: 134972,
                asOrganization: "KIDC LIMITED",
                country: "Japan",
                city: "Shibuya City",
                countryCode: "JP",
              },
            },
          ],
        },
      ],
      report: {
        sections: [
          {
            kind: "topology_table",
            rows: [
              {
                nodeName: "HK Node 01",
                proxyType: "shadowsocks",
                address: "z.h.s.1.p.i.b.e.d.hk05-ae5.entry.v51124-5.qpon",
                inboundASN: 4811,
                inboundIP: "114.28.148.163",
                inboundInfo: "Hongkou | China Telecom group",
                outboundASN: 134972,
                outboundIP: "103.151.173.98",
                outboundInfo: "Shibuya City | KIDC LIMITED",
                error: "",
              },
            ],
            columns: [],
          },
        ],
      },
    } as const;

    const sections = buildExportSections(archive as any, {
      hideCNInboundInExport: true,
    });

    const topologySection = sections.find((section) => section.kind === "topology_table");
    expect(topologySection).toBeTruthy();
    expect(topologySection?.rows[0]?.inboundASN).toBe("******");
    expect(topologySection?.rows[0]?.inboundIP).toBe("******");
    expect(topologySection?.rows[0]?.inboundInfo).toBe("******");
  });

  it("adds country-city to topology info when exporting older report rows", () => {
    const archive = {
      version: 1,
      task: {
        id: "task-1",
        vendor: "mihomo",
        context: {
          profileName: "Example",
          preset: "geo",
        },
      },
      state: {},
      results: [
        {
          proxyInfo: {
            name: "JP Node 01",
            type: "shadowsocks",
            address: "example.jp",
          },
          matrices: [
            {
              type: "GEOIP_INBOUND",
              payload: {
                ip: "114.28.148.163",
                asn: 4811,
                asOrganization: "Private Customer",
                isp: "China Telecom group",
                country: "China",
                city: "Hongkou",
                countryCode: "CN",
              },
            },
            {
              type: "GEOIP_OUTBOUND",
              payload: {
                ip: "103.151.173.98",
                asn: 134972,
                asOrganization: "KIDC LIMITED",
                country: "Japan",
                city: "Shibuya City",
                countryCode: "JP",
              },
            },
          ],
        },
      ],
      report: {
        sections: [
          {
            kind: "topology_table",
            rows: [
              {
                nodeName: "JP Node 01",
                proxyType: "shadowsocks",
                address: "example.jp",
                inboundASN: 4811,
                inboundIP: "114.28.148.163",
                inboundInfo: "Hongkou | China Telecom group",
                outboundASN: 134972,
                outboundIP: "103.151.173.98",
                outboundInfo: "Shibuya City | KIDC LIMITED",
                error: "",
              },
            ],
            columns: [],
          },
        ],
      },
    } as const;

    const sections = buildExportSections(archive as any);
    const topologySection = sections.find((section) => section.kind === "topology_table");
    expect(topologySection?.rows[0]?.inboundInfo).toBe("China-Hongkou | China Telecom group");
    expect(topologySection?.rows[0]?.outboundInfo).toBe("Japan-Shibuya City | KIDC LIMITED");
  });
});
