// automatically generated by the FlatBuffers compiler, do not modify

namespace gve_msg
{

using System;
using FlatBuffers;

public sealed class Attr : Table {
  public static Attr GetRootAsAttr(ByteBuffer _bb) { return GetRootAsAttr(_bb, new Attr()); }
  public static Attr GetRootAsAttr(ByteBuffer _bb, Attr obj) { return (obj.__init(_bb.GetInt(_bb.Position) + _bb.Position, _bb)); }
  public Attr __init(int _i, ByteBuffer _bb) { bb_pos = _i; bb = _bb; return this; }

  public float Atk { get { int o = __offset(4); return o != 0 ? bb.GetFloat(o + bb_pos) : (float)0.0f; } }
  public float Def { get { int o = __offset(6); return o != 0 ? bb.GetFloat(o + bb_pos) : (float)0.0f; } }
  public float Hp { get { int o = __offset(8); return o != 0 ? bb.GetFloat(o + bb_pos) : (float)0.0f; } }
  public float CritRate { get { int o = __offset(10); return o != 0 ? bb.GetFloat(o + bb_pos) : (float)0.0f; } }
  public float ResilienceRate { get { int o = __offset(12); return o != 0 ? bb.GetFloat(o + bb_pos) : (float)0.0f; } }
  public float CritValue { get { int o = __offset(14); return o != 0 ? bb.GetFloat(o + bb_pos) : (float)0.0f; } }
  public float ResilienceValue { get { int o = __offset(16); return o != 0 ? bb.GetFloat(o + bb_pos) : (float)0.0f; } }
  public int IceDamage { get { int o = __offset(18); return o != 0 ? bb.GetInt(o + bb_pos) : (int)0; } }
  public int IceDefense { get { int o = __offset(20); return o != 0 ? bb.GetInt(o + bb_pos) : (int)0; } }
  public float IceBonus { get { int o = __offset(22); return o != 0 ? bb.GetFloat(o + bb_pos) : (float)0.0f; } }
  public float IceResist { get { int o = __offset(24); return o != 0 ? bb.GetFloat(o + bb_pos) : (float)0.0f; } }
  public int FireDamage { get { int o = __offset(26); return o != 0 ? bb.GetInt(o + bb_pos) : (int)0; } }
  public int FireDefense { get { int o = __offset(28); return o != 0 ? bb.GetInt(o + bb_pos) : (int)0; } }
  public float FireBonus { get { int o = __offset(30); return o != 0 ? bb.GetFloat(o + bb_pos) : (float)0.0f; } }
  public float FireResist { get { int o = __offset(32); return o != 0 ? bb.GetFloat(o + bb_pos) : (float)0.0f; } }
  public int LightingDamage { get { int o = __offset(34); return o != 0 ? bb.GetInt(o + bb_pos) : (int)0; } }
  public int LightingDefense { get { int o = __offset(36); return o != 0 ? bb.GetInt(o + bb_pos) : (int)0; } }
  public float LightingBonus { get { int o = __offset(38); return o != 0 ? bb.GetFloat(o + bb_pos) : (float)0.0f; } }
  public float LightingResist { get { int o = __offset(40); return o != 0 ? bb.GetFloat(o + bb_pos) : (float)0.0f; } }
  public int PoisonDamage { get { int o = __offset(42); return o != 0 ? bb.GetInt(o + bb_pos) : (int)0; } }
  public int PoisonDefense { get { int o = __offset(44); return o != 0 ? bb.GetInt(o + bb_pos) : (int)0; } }
  public float PoisonBonus { get { int o = __offset(46); return o != 0 ? bb.GetFloat(o + bb_pos) : (float)0.0f; } }
  public float PoisonResist { get { int o = __offset(48); return o != 0 ? bb.GetFloat(o + bb_pos) : (float)0.0f; } }

  public static Offset<Attr> CreateAttr(FlatBufferBuilder builder,
      float atk = 0.0f,
      float def = 0.0f,
      float hp = 0.0f,
      float critRate = 0.0f,
      float resilienceRate = 0.0f,
      float critValue = 0.0f,
      float resilienceValue = 0.0f,
      int iceDamage = 0,
      int iceDefense = 0,
      float iceBonus = 0.0f,
      float iceResist = 0.0f,
      int fireDamage = 0,
      int fireDefense = 0,
      float fireBonus = 0.0f,
      float fireResist = 0.0f,
      int lightingDamage = 0,
      int lightingDefense = 0,
      float lightingBonus = 0.0f,
      float lightingResist = 0.0f,
      int poisonDamage = 0,
      int poisonDefense = 0,
      float poisonBonus = 0.0f,
      float poisonResist = 0.0f) {
    builder.StartObject(23);
    Attr.AddPoisonResist(builder, poisonResist);
    Attr.AddPoisonBonus(builder, poisonBonus);
    Attr.AddPoisonDefense(builder, poisonDefense);
    Attr.AddPoisonDamage(builder, poisonDamage);
    Attr.AddLightingResist(builder, lightingResist);
    Attr.AddLightingBonus(builder, lightingBonus);
    Attr.AddLightingDefense(builder, lightingDefense);
    Attr.AddLightingDamage(builder, lightingDamage);
    Attr.AddFireResist(builder, fireResist);
    Attr.AddFireBonus(builder, fireBonus);
    Attr.AddFireDefense(builder, fireDefense);
    Attr.AddFireDamage(builder, fireDamage);
    Attr.AddIceResist(builder, iceResist);
    Attr.AddIceBonus(builder, iceBonus);
    Attr.AddIceDefense(builder, iceDefense);
    Attr.AddIceDamage(builder, iceDamage);
    Attr.AddResilienceValue(builder, resilienceValue);
    Attr.AddCritValue(builder, critValue);
    Attr.AddResilienceRate(builder, resilienceRate);
    Attr.AddCritRate(builder, critRate);
    Attr.AddHp(builder, hp);
    Attr.AddDef(builder, def);
    Attr.AddAtk(builder, atk);
    return Attr.EndAttr(builder);
  }

  public static void StartAttr(FlatBufferBuilder builder) { builder.StartObject(23); }
  public static void AddAtk(FlatBufferBuilder builder, float atk) { builder.AddFloat(0, atk, 0.0f); }
  public static void AddDef(FlatBufferBuilder builder, float def) { builder.AddFloat(1, def, 0.0f); }
  public static void AddHp(FlatBufferBuilder builder, float hp) { builder.AddFloat(2, hp, 0.0f); }
  public static void AddCritRate(FlatBufferBuilder builder, float critRate) { builder.AddFloat(3, critRate, 0.0f); }
  public static void AddResilienceRate(FlatBufferBuilder builder, float resilienceRate) { builder.AddFloat(4, resilienceRate, 0.0f); }
  public static void AddCritValue(FlatBufferBuilder builder, float critValue) { builder.AddFloat(5, critValue, 0.0f); }
  public static void AddResilienceValue(FlatBufferBuilder builder, float resilienceValue) { builder.AddFloat(6, resilienceValue, 0.0f); }
  public static void AddIceDamage(FlatBufferBuilder builder, int iceDamage) { builder.AddInt(7, iceDamage, 0); }
  public static void AddIceDefense(FlatBufferBuilder builder, int iceDefense) { builder.AddInt(8, iceDefense, 0); }
  public static void AddIceBonus(FlatBufferBuilder builder, float iceBonus) { builder.AddFloat(9, iceBonus, 0.0f); }
  public static void AddIceResist(FlatBufferBuilder builder, float iceResist) { builder.AddFloat(10, iceResist, 0.0f); }
  public static void AddFireDamage(FlatBufferBuilder builder, int fireDamage) { builder.AddInt(11, fireDamage, 0); }
  public static void AddFireDefense(FlatBufferBuilder builder, int fireDefense) { builder.AddInt(12, fireDefense, 0); }
  public static void AddFireBonus(FlatBufferBuilder builder, float fireBonus) { builder.AddFloat(13, fireBonus, 0.0f); }
  public static void AddFireResist(FlatBufferBuilder builder, float fireResist) { builder.AddFloat(14, fireResist, 0.0f); }
  public static void AddLightingDamage(FlatBufferBuilder builder, int lightingDamage) { builder.AddInt(15, lightingDamage, 0); }
  public static void AddLightingDefense(FlatBufferBuilder builder, int lightingDefense) { builder.AddInt(16, lightingDefense, 0); }
  public static void AddLightingBonus(FlatBufferBuilder builder, float lightingBonus) { builder.AddFloat(17, lightingBonus, 0.0f); }
  public static void AddLightingResist(FlatBufferBuilder builder, float lightingResist) { builder.AddFloat(18, lightingResist, 0.0f); }
  public static void AddPoisonDamage(FlatBufferBuilder builder, int poisonDamage) { builder.AddInt(19, poisonDamage, 0); }
  public static void AddPoisonDefense(FlatBufferBuilder builder, int poisonDefense) { builder.AddInt(20, poisonDefense, 0); }
  public static void AddPoisonBonus(FlatBufferBuilder builder, float poisonBonus) { builder.AddFloat(21, poisonBonus, 0.0f); }
  public static void AddPoisonResist(FlatBufferBuilder builder, float poisonResist) { builder.AddFloat(22, poisonResist, 0.0f); }
  public static Offset<Attr> EndAttr(FlatBufferBuilder builder) {
    int o = builder.EndObject();
    return new Offset<Attr>(o);
  }
};


}
